package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

const (
	// nodeAutoprovisioningPollInterval is how often the cluster is polled while
	// Karpenter is installed or removed
	nodeAutoprovisioningPollInterval = 5 * time.Second
	// nodeAutoprovisioningTimeout bounds the convergence wait, the context of a
	// Terraform operation carries no deadline so the bound has to be explicit
	nodeAutoprovisioningTimeout = 15 * time.Minute
	// nodeAutoprovisioningPatchTimeout bounds the retry loop on locked features,
	// they stay locked while the cluster is not ACTIVE
	nodeAutoprovisioningPatchTimeout = 10 * time.Minute
	// nodeAutoprovisioningTransportRetries bounds how many answers in a row may
	// carry no HTTP status at all. A blip is worth another try, a dead endpoint
	// is not worth the whole budget
	nodeAutoprovisioningTransportRetries = 3
)

// nodeAutoprovisioningRequestTimeout bounds a single API call inside a loop's
// budget, so one stalled request cannot consume it. The client is built on
// http.DefaultClient, which carries no timeout of its own. It is a var so tests
// can shorten it
var nodeAutoprovisioningRequestTimeout = 30 * time.Second

func (r *ResourceKubernetes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// remove when going to beta public
	resp.Diagnostics.AddWarning(
		"Did you request product activation ?",
		"this product is not yet public and you need a support ticket to enable it on your organisation",
	)

	// remove when GA
	resp.Diagnostics.AddWarning(
		"Kubernetes product support is in beta",
		"It can break at any time, use it at your own risks",
	)

	plan := helper.From[Kubernetes](ctx, req.Plan, &resp.Diagnostics)
	config := helper.From[Kubernetes](ctx, req.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeAutoprovisioning, resolved := nodeAutoprovisioningValue(plan.NodeAutoprovisioning, config.NodeAutoprovisioning)
	if !resolved {
		resp.Diagnostics.AddError(unresolvedNodeAutoprovisioningSummary, unresolvedNodeAutoprovisioningDetail)
		return
	}
	enabled := nodeAutoprovisioning.ValueBool()

	// Create Kubernetes cluster
	createReq := tmp.KubernetesCreateRequest{
		Name: plan.Name.ValueString(),
		//KubeMajorVersion: "1.34",
	}
	// only send the features object when something is requested, so clusters
	// without node autoprovisioning keep the exact same payload
	if enabled {
		createReq.Features = &tmp.KubernetesFeatures{NodeAutoprovisioning: &enabled}
	}

	createRes := tmp.CreateKubernetes(ctx, r.Client(), r.Organization(), createReq)
	if createRes.HasError() {
		resp.Diagnostics.AddError("failed to create kubernetes cluster", createRes.Error().Error())
		return
	}
	k8sCluster := createRes.Payload()

	identity := KubernetesIdentity{
		ID: pkg.FromStr(k8sCluster.ID),
	}
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)

	// the requested value, not the one echoed back: features are reported by
	// proof of installation, the wait below makes it true
	state := Kubernetes{
		ID:                   pkg.FromStr(k8sCluster.ID),
		Name:                 pkg.FromStr(k8sCluster.Name),
		NodeAutoprovisioning: nodeAutoprovisioning,
	}

	for k8sCluster := range WaitForKubernetes(ctx, r.Client(), r.Organization(), k8sCluster.ID, 1*time.Second) {
		tflog.Info(ctx, "cluster state changed", map[string]any{"state": k8sCluster.Status})
		if k8sCluster.Status == "FAILED" {
			resp.Diagnostics.AddError("failed to provision kubernetes cluster", k8sCluster.Status)
			return // without running k8s cluster, nothing possible
		}
	}

	// no early return on failure here, the cluster exists and has to be saved in
	// state whatever Karpenter did, otherwise Terraform loses track of it. The
	// wait reports a failure as an error, so the apply fails instead of claiming
	// a feature the API never confirmed
	if enabled && !waitForNodeAutoprovisioning(
		ctx, r.Client(), r.Organization(), state.ID.ValueString(), true,
		nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
	) {
		resp.Diagnostics.AddWarning(
			"the cluster was created and is saved in state",
			"the failure happened after the cluster itself was created, so Terraform marks this resource tainted and the next apply would destroy and recreate it. "+
				"If Karpenter only needed more time, run terraform untaint on this resource and plan again rather than replacing the cluster",
		)
	}

	kubeConfigRes := tmp.GetKubeconfig(ctx, r.Client(), r.Organization(), k8sCluster.ID)
	if kubeConfigRes.HasError() {
		resp.Diagnostics.AddWarning("failed to get kubeconfig", kubeConfigRes.Error().Error())
	} else {
		kubeconfig := kubeConfigRes.Payload()
		state.KubeConfig = pkg.FromStr(string(*kubeconfig))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceKubernetes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	identity := helper.From[KubernetesIdentity](ctx, req.Identity, &resp.Diagnostics)
	state := Kubernetes{}

	if identity.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	kubernetesRes := tmp.GetKubernetes(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("Failed to get kubernetes instance", kubernetesRes.Error().Error())
	}

	k8sInfo := kubernetesRes.Payload()
	state.ID = identity.ID
	state.Name = pkg.FromStr(k8sInfo.Name)
	state.NodeAutoprovisioning = pkg.FromBool(autoprovisioningFeatureEnabled(k8sInfo.Features))

	// Get kubeconfig
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if !kubeConfigRes.HasError() {
		kubeconfig := kubeConfigRes.Payload()
		state.KubeConfig = pkg.FromStr(string(*kubeconfig))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceKubernetes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.From[Kubernetes](ctx, req.Plan, &resp.Diagnostics)
	config := helper.From[Kubernetes](ctx, req.Config, &resp.Diagnostics)
	state := helper.From[Kubernetes](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	nodeAutoprovisioning, resolved := nodeAutoprovisioningValue(plan.NodeAutoprovisioning, config.NodeAutoprovisioning)
	if !resolved {
		resp.Diagnostics.AddError(unresolvedNodeAutoprovisioningSummary, unresolvedNodeAutoprovisioningDetail)
		return
	}

	// name is still not patchable, only the features are
	if nodeAutoprovisioningChanged(nodeAutoprovisioning, state.NodeAutoprovisioning) {
		enabled := nodeAutoprovisioning.ValueBool()

		if !patchNodeAutoprovisioning(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), enabled,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningPatchTimeout, &resp.Diagnostics,
		) {
			return // keep the previous state, the patch is idempotent and will be replayed
		}

		if !waitForNodeAutoprovisioning(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), enabled,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
		) {
			// keep the previous state: the value was never observed, so the next
			// plan shows the change again and replays the idempotent patch
			return
		}
	}

	newState := *state
	newState.NodeAutoprovisioning = nodeAutoprovisioning

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *ResourceKubernetes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	identity := helper.From[KubernetesIdentity](ctx, req.Identity, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := tmp.DeleteKubernetes(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete kubernetes cluster", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// WaitForKubernetes monitors a Kubernetes cluster status and returns a channel
// that emits cluster object on the first call and whenever the Status field changes.
// The channel is automatically closed when status becomes ACTIVE or FAILED (terminal states).
func WaitForKubernetes(ctx context.Context, cc *client.Client, organisationID, clusterID string, pollInterval time.Duration) <-chan *tmp.ClusterView {
	ch := make(chan *tmp.ClusterView)

	go func() {
		var previousStatus string
		defer close(ch)

		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				res := tmp.GetKubernetesCluster(ctx, cc, organisationID, clusterID)
				if res.HasError() {
					continue
				}
				status := res.Payload().Status

				if status != previousStatus {
					ch <- res.Payload()
					previousStatus = status
				}

				if status == "ACTIVE" || status == "FAILED" {
					return
				}
			}
		}
	}()

	return ch
}

// autoprovisioningFeatureEnabled reports whether the API says node autoprovisioning is installed.
// The API omits the features object on clusters that never used it, which means
// the same thing as disabled
func autoprovisioningFeatureEnabled(features *tmp.KubernetesFeatures) bool {
	return features != nil && features.NodeAutoprovisioning != nil && *features.NodeAutoprovisioning
}

const (
	unresolvedNodeAutoprovisioningSummary = "cannot resolve node_autoprovisioning"
	unresolvedNodeAutoprovisioningDetail  = "the value is still unknown at apply time, which Terraform is supposed to make impossible: " +
		"the provider cannot tell whether Karpenter should be installed. Please report this with the configuration that produced it"
)

// nodeAutoprovisioningValue resolves the value to apply and to save in state.
// The schema default only fills a null *configuration* value, so an expression
// that could not be resolved at plan time leaves the planned value unknown, and
// ValueBool would silently answer false: enabling would become a no-op, and a
// change would become a request to uninstall Karpenter. The configuration is
// resolved by the time apply runs, so it is the source of truth whenever the
// planned value is still unknown
func nodeAutoprovisioningValue(plan, config types.Bool) (types.Bool, bool) {
	if !plan.IsUnknown() {
		return plan, true
	}
	if config.IsUnknown() {
		return plan, false
	}
	if config.IsNull() {
		// the attribute is absent from the configuration, the schema default
		// applies. The framework normally fills it before we get here, this is
		// the belt to that pair of braces
		return types.BoolValue(nodeAutoprovisioningDefault), true
	}

	return config, true
}

// pollErrorFatal reports whether an error returned while polling the cluster is
// permanent. A 404 means the cluster is gone, a 401 or a 403 means the
// credentials stopped working: retrying any of those until the deadline hides
// the real cause behind a generic timeout. Everything else, a transport failure
// reported as status 0 included, is transient and worth another tick
func pollErrorFatal(statusCode int) bool {
	switch statusCode {
	case http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden:
		return true
	default:
		return false
	}
}

// pollErrorDetail explains a permanent polling error
func pollErrorDetail(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "the cluster does not exist any more, it was deleted while Karpenter was being installed or removed"
	case http.StatusUnauthorized, http.StatusForbidden:
		return "the credentials were refused while polling the cluster, they may have expired or lost access to this organisation"
	default:
		return ""
	}
}

// clusterTerminalStatus reports whether the cluster reached a status the feature
// can never converge from. It has to be checked before convergence itself: a
// deleted cluster reports no features at all, which a disable request would
// otherwise read as proof that Karpenter was successfully removed
func clusterTerminalStatus(status string) bool {
	switch status {
	case "FAILED", "DELETING", "DELETED":
		return true
	default:
		return false
	}
}

// classifyPatchError maps the HTTP status of a features patch to a retry decision
// and a user facing explanation. Features are locked while the cluster is not
// ACTIVE, which the API signals with a 412 and resolves on its own, while a 400
// and a 409 both need the user to change something first
func classifyPatchError(statusCode int, enabled bool) (retryable bool, detail string) {
	switch statusCode {
	case http.StatusPreconditionFailed:
		return true, ""
	case http.StatusBadRequest:
		// the cluster wide autoscalingEnabled feature is mutually exclusive with
		// this one and is not exposed by this provider
		return false, "node autoprovisioning cannot run alongside node group autoscaling, disable the autoscalingEnabled feature of the cluster first"
	case http.StatusConflict:
		if enabled {
			return false, "a Karpenter installation already runs in the kube-system namespace of this cluster, remove it before enabling node_autoprovisioning"
		}
		return false, "Karpenter custom resources still exist in this cluster, delete every NodePool, NodeClaim, NodeOverlay and CleverNodeClass, let Karpenter drain the nodes, then retry"
	case http.StatusNotFound:
		return false, "the cluster does not exist any more"
	case http.StatusUnauthorized, http.StatusForbidden:
		return false, "the credentials were refused, they may have expired or lost access to this organisation"
	case 0:
		// no HTTP status at all: the request failed before a response could be
		// read, so whether the API applied it is unknowable from here. The patch
		// carries an explicit boolean and is idempotent, so sending it again is
		// safe and is the only way to find out
		return true, ""
	}

	// a status we did receive but could not read is not evidence of a permanent
	// rejection: it is the body read that failed, not the patch
	if statusCode >= 200 && statusCode < 300 {
		return true, ""
	}

	return false, ""
}

// patchExpiryDetail explains a patch loop that ran out of budget, in terms of
// whatever it was still retrying. Only a 412 licenses the claim that the
// cluster's features were locked
func patchExpiryDetail(lastStatus int, timeout time.Duration) string {
	switch lastStatus {
	case http.StatusPreconditionFailed:
		return fmt.Sprintf("the cluster never accepted the change in %s, its features stay locked while it is not ACTIVE", timeout)
	case 0:
		return fmt.Sprintf("the Clever Cloud API could not be reached for %s, the change may or may not have been applied", timeout)
	default:
		return fmt.Sprintf("the cluster never accepted the change in %s", timeout)
	}
}

// terminalStatusDetail explains a status the feature can never converge from. A
// failed cluster is resumed on the Clever Cloud side, a cluster on its way out
// cannot be resumed at all
func terminalStatusDetail(status string) string {
	if status == "FAILED" {
		return "the cluster went FAILED while node autoprovisioning was being installed or removed, it has to be resumed before it accepts any other change"
	}

	return fmt.Sprintf("the cluster went %s while node autoprovisioning was being installed or removed, it cannot be resumed: recreate the cluster or remove it from state", status)
}

// nodeAutoprovisioningChanged reports whether the cluster has to be patched.
// Values are compared unwrapped: a state written before this attribute existed
// holds null, which means the same thing as disabled and must not be patched
func nodeAutoprovisioningChanged(plan, state types.Bool) bool {
	return plan.ValueBool() != state.ValueBool()
}

// patchErrorDetail assembles the detail of a features patch diagnostic, keeping
// only the parts the API actually gave us
func patchErrorDetail(explanation, apiError, requestID string) string {
	parts := []string{}
	if explanation != "" {
		parts = append(parts, explanation)
	}
	if apiError != "" {
		parts = append(parts, apiError)
	}
	if requestID != "" {
		parts = append(parts, fmt.Sprintf("request id: %s", requestID))
	}

	return strings.Join(parts, "\n")
}

// loopFailureDetail explains why a bounded loop gave up, telling an apply the
// user interrupted apart from a deadline that expired on its own. lastErr is
// appended when the caller still has an error that was current at that point
func loopFailureDetail(parent context.Context, expired string, lastErr string) string {
	if parent.Err() != nil {
		return patchErrorDetail("the apply was interrupted before the change could be confirmed", parent.Err().Error(), "")
	}

	return patchErrorDetail(expired, lastErr, "")
}

// patchNodeAutoprovisioning enables or disables node autoprovisioning on a
// cluster. Features are locked while the cluster is not ACTIVE, so the patch is
// retried until the cluster accepts it
func patchNodeAutoprovisioning(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	enabled bool,
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) bool {
	summary := "failed to disable node autoprovisioning"
	if enabled {
		summary = "failed to enable node autoprovisioning"
	}

	// the Terraform operation context carries no deadline and neither does the
	// API client, so the budget has to be a context or a stalled request would
	// outlive the timeout this loop believes it enforces
	loopCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	lastStatus, lastErr, transportFailures := 0, "", 0

	for {
		reqCtx, cancelRequest := context.WithTimeout(loopCtx, nodeAutoprovisioningRequestTimeout)
		res := tmp.UpdateKubernetes(reqCtx, cc, organisationID, clusterID, tmp.KubernetesPatchRequest{
			Features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &enabled},
		})
		cancelRequest()

		if !res.HasError() {
			return true
		}

		// a failure our own budget caused says nothing about the API: leave the
		// last real answer standing as the reason and fall through to the
		// deadline check below, which is what actually happened
		if loopCtx.Err() == nil {
			lastStatus, lastErr = res.StatusCode(), res.Error().Error()

			retryable, detail := classifyPatchError(lastStatus, enabled)
			if !retryable {
				diags.AddError(summary, patchErrorDetail(detail, lastErr, res.SozuID()))
				return false
			}

			// an answer carrying no status at all proves nothing about the API
			// being there; a few in a row mean it is not
			if lastStatus == 0 {
				transportFailures++
				if transportFailures >= nodeAutoprovisioningTransportRetries {
					diags.AddError(summary, patchErrorDetail(
						fmt.Sprintf("the Clever Cloud API could not be reached in %d attempts, the change may or may not have been applied", transportFailures),
						lastErr, res.SozuID(),
					))

					return false
				}
			} else {
				transportFailures = 0
			}

			tflog.Warn(ctx, "features patch was not accepted, retrying", map[string]any{"status": lastStatus})
		}

		if loopCtx.Err() != nil {
			diags.AddError(summary, loopFailureDetail(ctx, patchExpiryDetail(lastStatus, timeout), lastErr))
			return false
		}

		select {
		case <-loopCtx.Done():
			diags.AddError(summary, loopFailureDetail(ctx, patchExpiryDetail(lastStatus, timeout), lastErr))
			return false
		case <-ticker.C:
		}
	}
}

// waitForNodeAutoprovisioning polls the cluster until the API reports the
// expected node autoprovisioning value, and reports whether it ever did.
// Features are reported by proof of installation: right after the API accepts
// the change it keeps returning the previous value until Karpenter is really
// installed or removed, so the value has to converge before it can be saved in
// state. Giving up is an error, never a silent success: the whole point of the
// wait is that the value in state was actually observed
func waitForNodeAutoprovisioning(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	expected bool,
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) bool {
	summary := "node autoprovisioning was not disabled"
	if expected {
		summary = "node autoprovisioning was not enabled"
	}

	loopCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	expired := fmt.Sprintf("the API never reported the requested value in %s, Karpenter may still be installing or being removed, run terraform plan once the cluster is ACTIVE to reconcile it", timeout)
	lastErr := ""

	for {
		// probe before waiting, a ticker never fires at t=0
		reqCtx, cancelRequest := context.WithTimeout(loopCtx, nodeAutoprovisioningRequestTimeout)
		res := tmp.GetKubernetesCluster(reqCtx, cc, organisationID, clusterID)
		cancelRequest()

		switch {
		case res.HasError():
			// a failure our own budget caused says nothing about the cluster,
			// the deadline check below is what actually happened
			if loopCtx.Err() != nil {
				break
			}

			lastErr = res.Error().Error()
			if pollErrorFatal(res.StatusCode()) {
				diags.AddError(summary, patchErrorDetail(pollErrorDetail(res.StatusCode()), lastErr, res.SozuID()))
				return false
			}
			tflog.Warn(ctx, "failed to poll cluster features", map[string]any{"error": lastErr})
		case clusterTerminalStatus(res.Payload().Status):
			// installing or removing Karpenter can fail the whole cluster, and
			// the cluster can also be deleted from under us: neither ever
			// converges, and the two need different advice
			diags.AddError(summary, terminalStatusDetail(res.Payload().Status))
			return false
		case autoprovisioningFeatureEnabled(res.Payload().Features) == expected:
			tflog.Info(ctx, "node autoprovisioning converged", map[string]any{"enabled": expected})
			return true
		default:
			// the cluster answered, so whatever failed before has recovered and
			// must not be reported as the reason the loop later gave up
			lastErr = ""
			tflog.Debug(ctx, "waiting for node autoprovisioning", map[string]any{
				"status":  res.Payload().Status,
				"enabled": autoprovisioningFeatureEnabled(res.Payload().Features),
			})
		}

		if loopCtx.Err() != nil {
			diags.AddError(summary, loopFailureDetail(ctx, expired, lastErr))
			return false
		}

		select {
		case <-loopCtx.Done():
			diags.AddError(summary, loopFailureDetail(ctx, expired, lastErr))
			return false
		case <-ticker.C:
		}
	}
}
