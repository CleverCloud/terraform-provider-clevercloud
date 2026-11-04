package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
	"go.clever-cloud.dev/sdk/models"
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
	// versionUpgradeTimeout bounds a version upgrade, which rolls the control
	// plane and then every node group
	versionUpgradeTimeout = 2 * time.Hour
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
	persistentStorage := resolveValue(plan.PersistentStorage, config.PersistentStorage)
	tags := resolveValue(plan.Tags, config.Tags)

	// Create Kubernetes cluster
	createReq := createPayload(ctx, plan.Name.ValueString(), enabled, persistentStorage, tags, resolveValue(plan.Topology, config.Topology), resolveValue(plan.Version, config.Version), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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
		PersistentStorage:    persistentStorage,
		Tags:                 tags,
		Topology:             topologyFromAPI(ctx, k8sCluster.Topology, &resp.Diagnostics),
		Version:              pkg.FromStr(k8sCluster.Version),
	}
	if !isSet(state.PersistentStorage) {
		state.PersistentStorage = pkg.FromBool(persistentStorageFeatureEnabled(k8sCluster.Features))
	}

	for k8sCluster := range WaitForKubernetes(ctx, r.Client(), r.Organization(), k8sCluster.ID, 1*time.Second) {
		tflog.Info(ctx, "cluster state changed", map[string]any{"state": k8sCluster.Status})
		if k8sCluster.Status == "FAILED" {
			resp.Diagnostics.AddError("failed to provision kubernetes cluster", k8sCluster.Status)
			return // without running k8s cluster, nothing possible
		}
	}

	// no early return on failure here, the cluster exists and has to be saved in
	// state whatever the features did, otherwise Terraform loses track of it. The
	// waits report a failure as an error, so the apply fails instead of claiming
	// a feature the API never confirmed
	featuresConverged := true
	if enabled {
		featuresConverged = waitForNodeAutoprovisioning(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), true,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
		)
	}
	if persistentStorage.ValueBool() {
		featuresConverged = waitForPersistentStorage(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), true,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
		) && featuresConverged
	}
	if !featuresConverged {
		resp.Diagnostics.AddWarning(
			"the cluster was created and is saved in state",
			"the failure happened after the cluster itself was created, so Terraform marks this resource tainted and the next apply would destroy and recreate it. "+
				"If the feature only needed more time, run terraform untaint on this resource and plan again rather than replacing the cluster",
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
	identity := KubernetesIdentity{ID: helper.IdentityIDFrom(ctx, req.Identity, req.State, &resp.Diagnostics)}
	state := Kubernetes{}
	if resp.Diagnostics.HasError() {
		return
	}

	// The framework requires an identity even when Read removes a missing
	// cluster from state. Recover it before any removal return below.
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if identity.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	kubernetesRes := tmp.GetKubernetes(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if kubernetesRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("Failed to get kubernetes instance", kubernetesRes.Error().Error())
		return
	}

	// the state is rebuilt from the API, but whether tags are managed is only
	// known from the previous one. It holds nothing but the id on an import
	var previousTags types.Set
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tags"), &previousTags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	k8sInfo := kubernetesRes.Payload()
	state.ID = identity.ID
	state.Name = pkg.FromStr(k8sInfo.Name)
	state.NodeAutoprovisioning = pkg.FromBool(autoprovisioningFeatureEnabled(k8sInfo.Features))
	state.PersistentStorage = pkg.FromBool(persistentStorageFeatureEnabled(k8sInfo.Features))
	state.Tags = resources.TagsFromAPI(ctx, k8sInfo.Tags, previousTags, &resp.Diagnostics)
	state.Topology = topologyFromAPI(ctx, k8sInfo.Topology, &resp.Diagnostics)
	state.Version = pkg.FromStr(k8sInfo.Version)

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

	persistentStorage := resolveValue(plan.PersistentStorage, config.PersistentStorage)
	tags := resolveValue(plan.Tags, config.Tags)
	topology := resolveValue(plan.Topology, config.Topology)
	version := resolveValue(plan.Version, config.Version)
	clusterID := state.ID.ValueString()

	// name and tags are applied on the spot, they do not reconcile the cluster
	if payload := metadataPatch(ctx, plan.Name, state.Name, tags, state.Tags, &resp.Diagnostics); payload != nil && !patchCluster(
		ctx, r.Client(), r.Organization(), clusterID, *payload, "failed to update the kubernetes cluster", classifyClusterPatchError,
		nodeAutoprovisioningPollInterval, nodeAutoprovisioningPatchTimeout, &resp.Diagnostics,
	) {
		return // keep the previous state, the patch is idempotent and will be replayed
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if isSet(persistentStorage) && persistentStorage.ValueBool() != state.PersistentStorage.ValueBool() {
		enabled := persistentStorage.ValueBool()

		if !patchPersistentStorage(
			ctx, r.Client(), r.Organization(), clusterID, enabled,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningPatchTimeout, &resp.Diagnostics,
		) {
			return // keep the previous state, the patch is idempotent and will be replayed
		}

		if !waitForPersistentStorage(
			ctx, r.Client(), r.Organization(), clusterID, enabled,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
		) {
			return // keep the previous state, the next plan shows the change again
		}
	}

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

	// last, as the API refuses a version change while a feature is being reconciled. A
	// state holding no version predates the attribute, the next refresh tells
	// what the cluster really runs
	if isSet(version) && isSet(state.Version) && !version.Equal(state.Version) && !changeClusterVersion(
		ctx, r.Client(), r.Organization(), clusterID, version.ValueString(),
		nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, versionUpgradeTimeout, &resp.Diagnostics,
	) {
		return // keep the previous state, the next plan shows the change again
	}

	newState := *state
	newState.Name = plan.Name
	newState.NodeAutoprovisioning = nodeAutoprovisioning
	newState.Tags = tags
	if isSet(persistentStorage) {
		newState.PersistentStorage = persistentStorage
	}
	if isSet(topology) {
		newState.Topology = topology
	}
	if isSet(version) {
		newState.Version = version
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, KubernetesIdentity{ID: state.ID})...)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *ResourceKubernetes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	identity := KubernetesIdentity{ID: helper.IdentityIDFrom(ctx, req.Identity, req.State, &resp.Diagnostics)}
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

// persistentStorageFeatureEnabled reports whether the API says the CSI driver is
// installed, an absent flag means disabled
func persistentStorageFeatureEnabled(features *tmp.KubernetesFeatures) bool {
	return features != nil && features.Csi != nil && *features.Csi
}

// resolveValue returns the value to apply. An expression that could not be resolved
// at plan time leaves the planned value unknown while the configuration knows it
// by the time apply runs, see nodeAutoprovisioningValue. The result is null when
// the attribute is absent from the configuration and the API picks the value
func resolveValue[T attr.Value](plan, config T) T {
	if plan.IsUnknown() {
		return config
	}

	return plan
}

// createPayload builds the cluster creation request. Attributes without a value
// are left out so the API applies its own defaults, and the features object is
// only sent when something is requested
func createPayload(
	ctx context.Context,
	name string,
	nodeAutoprovisioning bool,
	persistentStorage types.Bool,
	tags types.Set,
	topology types.Object,
	version types.String,
	diags *diag.Diagnostics,
) tmp.KubernetesCreateRequest {
	payload := tmp.KubernetesCreateRequest{
		Name:     name,
		Topology: topologyToAPI(ctx, topology, diags),
		Version:  version.ValueString(),
	}

	if isSet(tags) {
		payload.Tags = pkg.SetToStringSlice(ctx, tags, diags)
	}

	if nodeAutoprovisioning || persistentStorage.ValueBool() {
		payload.Features = &tmp.KubernetesFeatures{}
		if nodeAutoprovisioning {
			payload.Features.NodeAutoprovisioning = &nodeAutoprovisioning
		}
		if persistentStorage.ValueBool() {
			enabled := true
			payload.Features.Csi = &enabled
		}
	}

	return payload
}

// metadataPatch builds the request applying a new name and new tags, nil when
// neither changed. Tags without a planned value are not managed by Terraform and
// are left alone
func metadataPatch(ctx context.Context, name, stateName types.String, tags, stateTags types.Set, diags *diag.Diagnostics) *tmp.KubernetesPatchRequest {
	payload := tmp.KubernetesPatchRequest{}
	changed := false

	if !name.Equal(stateName) {
		newName := name.ValueString()
		payload.Name = &newName
		changed = true
	}

	if isSet(tags) && !tags.Equal(stateTags) {
		newTags := pkg.SetToStringSlice(ctx, tags, diags)
		if newTags == nil {
			newTags = []string{} // null would not remove them
		}
		payload.Tags = &newTags
		changed = true
	}

	if !changed {
		return nil
	}

	return &payload
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
		return "the cluster does not exist any more, it was deleted while Terraform was waiting for it"
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
	case http.StatusBadRequest:
		// the cluster wide autoscalingEnabled feature is mutually exclusive with
		// this one and is not exposed by this provider
		return false, "node autoprovisioning cannot run alongside the cluster wide autoscalingEnabled feature, disable it first"
	case http.StatusConflict:
		if enabled {
			return false, "a Karpenter installation already runs in the kube-system namespace of this cluster, remove it before enabling node_autoprovisioning"
		}
		return false, "Karpenter custom resources still exist in this cluster, delete every NodePool, NodeClaim, NodeOverlay and CleverNodeClass, let Karpenter drain the nodes, then retry"
	}

	return classifyClusterPatchError(statusCode)
}

// classifyClusterPatchError is the retry decision of a patch whose rejection
// needs no feature specific explanation, the API error is enough
func classifyClusterPatchError(statusCode int) (retryable bool, detail string) {
	switch statusCode {
	case http.StatusPreconditionFailed:
		return true, ""
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
	}

	// an unreadable 2xx is retried, so running out of budget on one means the
	// API kept accepting the change while none of its answers could be read
	if lastStatus >= 200 && lastStatus < 300 {
		return fmt.Sprintf("the API accepted the change but none of its responses could be read in %s, run terraform plan to check whether it was applied", timeout)
	}

	return fmt.Sprintf("the cluster never accepted the change in %s", timeout)
}

// terminalStatusDetail explains a status the cluster can never converge from,
// operation being what was going on. A failed cluster is resumed on the Clever
// Cloud side, a cluster on its way out cannot be resumed at all
func terminalStatusDetail(status, operation string) string {
	if status == "FAILED" {
		return fmt.Sprintf("the cluster went FAILED while %s, it has to be resumed before it accepts any other change", operation)
	}

	return fmt.Sprintf("the cluster went %s while %s, it cannot be resumed: recreate the cluster or remove it from state", status, operation)
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

	return patchCluster(ctx, cc, organisationID, clusterID, tmp.KubernetesPatchRequest{
		Features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &enabled},
	}, summary, func(statusCode int) (bool, string) {
		return classifyPatchError(statusCode, enabled)
	}, pollInterval, timeout, diags)
}

// patchPersistentStorage enables or disables the CSI driver of a cluster, with
// the same retries on locked features as patchNodeAutoprovisioning
func patchPersistentStorage(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	enabled bool,
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) bool {
	summary := "failed to disable persistent storage"
	if enabled {
		summary = "failed to enable persistent storage"
	}

	return patchCluster(ctx, cc, organisationID, clusterID, tmp.KubernetesPatchRequest{
		Features: &tmp.KubernetesFeatures{Csi: &enabled},
	}, summary, classifyClusterPatchError, pollInterval, timeout, diags)
}

// patchCluster sends an idempotent patch until the cluster accepts it. classify
// decides which answers are worth another try and explains the others
func patchCluster(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	payload tmp.KubernetesPatchRequest,
	summary string,
	classify func(statusCode int) (retryable bool, detail string),
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) bool {
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
		res := tmp.UpdateKubernetes(reqCtx, cc, organisationID, clusterID, payload)
		cancelRequest()

		if !res.HasError() {
			return true
		}

		// a failure our own budget caused says nothing about the API: leave the
		// last real answer standing as the reason and fall through to the
		// deadline check below, which is what actually happened
		if loopCtx.Err() == nil {
			lastStatus, lastErr = res.StatusCode(), res.Error().Error()

			retryable, detail := classify(lastStatus)
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

	return waitForCluster(ctx, cc, organisationID, clusterID, clusterWait{
		summary:   summary,
		operation: "node autoprovisioning was being installed or removed",
		expired:   fmt.Sprintf("the API never reported the requested value in %s, Karpenter may still be installing or being removed, run terraform plan once the cluster is ACTIVE to reconcile it", timeout),
		reached: func(cluster *tmp.ClusterView) bool {
			return autoprovisioningFeatureEnabled(cluster.Features) == expected
		},
	}, pollInterval, timeout, diags)
}

// waitForPersistentStorage polls the cluster until the API reports the expected
// CSI value on an ACTIVE cluster. Like node autoprovisioning it is reported by
// proof of installation, and the cluster reconciles while it is installed
func waitForPersistentStorage(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	expected bool,
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) bool {
	summary := "persistent storage was not disabled"
	if expected {
		summary = "persistent storage was not enabled"
	}

	return waitForCluster(ctx, cc, organisationID, clusterID, clusterWait{
		summary:   summary,
		operation: "persistent storage was being installed or removed",
		expired:   fmt.Sprintf("the API never reported the requested value in %s, run terraform plan once the cluster is ACTIVE to reconcile it", timeout),
		reached: func(cluster *tmp.ClusterView) bool {
			return cluster.Status == "ACTIVE" && persistentStorageFeatureEnabled(cluster.Features) == expected
		},
	}, pollInterval, timeout, diags)
}

// changeClusterVersion moves a cluster to another Kubernetes version, older ones
// included, and waits for the end of the rollout. The API answers a 409 unless the cluster is ACTIVE, and
// reports the target version as soon as it accepts the request, so ACTIVE is the
// only proof that the rollout is over
func changeClusterVersion(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID, version string,
	pollInterval, readyTimeout, upgradeTimeout time.Duration,
	diags *diag.Diagnostics,
) bool {
	const summary = "failed to change the kubernetes version"

	if !waitForCluster(ctx, cc, organisationID, clusterID, clusterWait{
		summary:   summary,
		operation: "waiting to change its version",
		expired:   fmt.Sprintf("the cluster did not become ACTIVE in %s, the version change was not started", readyTimeout),
		reached:   func(cluster *tmp.ClusterView) bool { return cluster.Status == "ACTIVE" },
	}, pollInterval, readyTimeout, diags) {
		return false
	}

	res := sdk.NewSDK(sdk.WithClient(cc)).V4().Kubernetes().Organisations().Ownerid(organisationID).Clusters().Kubernetesid(clusterID).
		Version().Update().Updatekubernetesclusterversion(ctx, &models.PatchClusterVersion{TargetVersion: version})
	if res.HasError() {
		diags.AddError(summary, patchErrorDetail("", res.Error().Error(), res.SozuID()))
		return false
	}

	return waitForCluster(ctx, cc, organisationID, clusterID, clusterWait{
		summary:   summary,
		operation: "its version was being changed",
		expired:   fmt.Sprintf("the version change did not finish in %s, it may still be running, run terraform plan once the cluster is ACTIVE to reconcile it", upgradeTimeout),
		reached:   func(cluster *tmp.ClusterView) bool { return cluster.Status == "ACTIVE" && cluster.Version == version },
	}, pollInterval, upgradeTimeout, diags)
}

// clusterWait describes what waitForCluster is waiting for
type clusterWait struct {
	summary   string                              // summary of the diagnostics
	operation string                              // what was going on, should the cluster fail
	expired   string                              // explanation of an expired timeout
	reached   func(cluster *tmp.ClusterView) bool // whether the wait is over
}

// waitForCluster polls the cluster until wait.reached accepts it, and reports
// whether it ever did. Giving up is an error, never a silent success
func waitForCluster(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	wait clusterWait,
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) bool {
	summary := wait.summary

	loopCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

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
			tflog.Warn(ctx, "failed to poll the cluster", map[string]any{"error": lastErr})
		case clusterTerminalStatus(res.Payload().Status):
			// the operation can fail the whole cluster, and the cluster can also
			// be deleted from under us: neither ever converges, and the two need
			// different advice
			diags.AddError(summary, terminalStatusDetail(res.Payload().Status, wait.operation))
			return false
		case wait.reached(res.Payload()):
			tflog.Info(ctx, "the cluster converged", map[string]any{"operation": wait.operation})
			return true
		default:
			// the cluster answered, so whatever failed before has recovered and
			// must not be reported as the reason the loop later gave up
			lastErr = ""
			tflog.Debug(ctx, "waiting for the cluster", map[string]any{"status": res.Payload().Status, "operation": wait.operation})
		}

		if loopCtx.Err() != nil {
			diags.AddError(summary, loopFailureDetail(ctx, wait.expired, lastErr))
			return false
		}

		select {
		case <-loopCtx.Done():
			diags.AddError(summary, loopFailureDetail(ctx, wait.expired, lastErr))
			return false
		case <-ticker.C:
		}
	}
}
