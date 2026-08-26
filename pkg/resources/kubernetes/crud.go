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
)

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
	if resp.Diagnostics.HasError() {
		return
	}

	// Create Kubernetes cluster
	createReq := tmp.KubernetesCreateRequest{
		Name: plan.Name.ValueString(),
		//KubeMajorVersion: "1.34",
	}
	// only send the features object when something is requested, so clusters
	// without node autoprovisioning keep the exact same payload
	if plan.NodeAutoprovisioning.ValueBool() {
		enabled := true
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
		NodeAutoprovisioning: plan.NodeAutoprovisioning,
	}

	for k8sCluster := range WaitForKubernetes(ctx, r.Client(), r.Organization(), k8sCluster.ID, 1*time.Second) {
		tflog.Info(ctx, "cluster state changed", map[string]any{"state": k8sCluster.Status})
		if k8sCluster.Status == "FAILED" {
			resp.Diagnostics.AddError("failed to provision kubernetes cluster", k8sCluster.Status)
			return // without running k8s cluster, nothing possible
		}
	}

	// no early return on failure here, the cluster exists and has to be saved in
	// state whatever Karpenter did, otherwise Terraform loses track of it
	if plan.NodeAutoprovisioning.ValueBool() {
		waitForNodeAutoprovisioning(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), true,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
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
	state.NodeAutoprovisioning = pkg.FromBool(featureEnabled(k8sInfo.Features))

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
	state := helper.From[Kubernetes](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// name is still not patchable, only the features are
	if nodeAutoprovisioningChanged(plan.NodeAutoprovisioning, state.NodeAutoprovisioning) {
		enabled := plan.NodeAutoprovisioning.ValueBool()

		if !patchNodeAutoprovisioning(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), enabled,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningPatchTimeout, &resp.Diagnostics,
		) {
			return // keep the previous state, the patch is idempotent and will be replayed
		}

		waitForNodeAutoprovisioning(
			ctx, r.Client(), r.Organization(), state.ID.ValueString(), enabled,
			nodeAutoprovisioningPollInterval, nodeAutoprovisioningTimeout, &resp.Diagnostics,
		)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	newState := *state
	newState.NodeAutoprovisioning = plan.NodeAutoprovisioning

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

// featureEnabled reports whether the API says node autoprovisioning is installed.
// The API omits the features object on clusters that never used it, which means
// the same thing as disabled
func featureEnabled(features *tmp.KubernetesFeatures) bool {
	return features != nil && features.NodeAutoprovisioning != nil && *features.NodeAutoprovisioning
}

// classifyPatchError maps the HTTP status of a features patch to a retry decision
// and a user facing explanation. Features are locked while the cluster is not
// ACTIVE, which the API signals with a 412 and resolves on its own, while a 409
// is a conflict inside the cluster only the user can solve
func classifyPatchError(statusCode int, enabled bool) (retryable bool, detail string) {
	switch statusCode {
	case http.StatusPreconditionFailed:
		return true, ""
	case http.StatusConflict:
		if enabled {
			return false, "a Karpenter installation already runs in the kube-system namespace of this cluster, remove it before enabling node_autoprovisioning"
		}
		return false, "Karpenter custom resources still exist in this cluster, delete every NodePool, NodeClaim, NodeOverlay and CleverNodeClass before disabling node_autoprovisioning"
	case 0:
		return false, "the request never reached the Clever Cloud API"
	default:
		return false, ""
	}
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

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		res := tmp.UpdateKubernetes(ctx, cc, organisationID, clusterID, tmp.KubernetesPatchRequest{
			Features: &tmp.KubernetesFeatures{NodeAutoprovisioning: &enabled},
		})
		if !res.HasError() {
			return true
		}

		retryable, detail := classifyPatchError(res.StatusCode(), enabled)
		if !retryable {
			diags.AddError(summary, patchErrorDetail(detail, res.Error().Error(), res.SozuID()))
			return false
		}

		tflog.Warn(ctx, "cluster features are locked, retrying", map[string]any{"status": res.StatusCode()})

		if time.Now().After(deadline) {
			diags.AddError(summary, fmt.Sprintf("the cluster features stayed locked for %s, the cluster never went back to ACTIVE", timeout))
			return false
		}

		select {
		case <-ctx.Done():
			diags.AddError(summary, ctx.Err().Error())
			return false
		case <-ticker.C:
		}
	}
}

// waitForNodeAutoprovisioning polls the cluster until the API reports the
// expected node autoprovisioning value. Features are reported by proof of
// installation: right after the API accepts the change it keeps returning the
// previous value until Karpenter is really installed or removed, so the value
// has to converge before it can be saved in state
func waitForNodeAutoprovisioning(
	ctx context.Context,
	cc *client.Client,
	organisationID, clusterID string,
	expected bool,
	pollInterval, timeout time.Duration,
	diags *diag.Diagnostics,
) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		// probe before waiting, a ticker never fires at t=0
		res := tmp.GetKubernetesCluster(ctx, cc, organisationID, clusterID)
		switch {
		case res.HasError():
			tflog.Warn(ctx, "failed to poll cluster features", map[string]any{"error": res.Error().Error()})
		case res.Payload().Status == "FAILED":
			diags.AddError("failed to provision kubernetes cluster", res.Payload().Status)
			return
		case featureEnabled(res.Payload().Features) == expected:
			tflog.Info(ctx, "node autoprovisioning converged", map[string]any{"enabled": expected})
			return
		default:
			tflog.Debug(ctx, "waiting for node autoprovisioning", map[string]any{
				"status":  res.Payload().Status,
				"enabled": featureEnabled(res.Payload().Features),
			})
		}

		if time.Now().After(deadline) {
			diags.AddWarning(
				"node autoprovisioning did not converge",
				fmt.Sprintf("Karpenter is still being installed or removed after %s, the value saved in state is the one requested, run terraform plan again once the cluster is ACTIVE to check it", timeout),
			)
			return
		}

		select {
		case <-ctx.Done():
			diags.AddWarning("node autoprovisioning wait cancelled", ctx.Err().Error())
			return
		case <-ticker.C:
		}
	}
}
