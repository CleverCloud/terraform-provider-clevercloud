package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func (r *ResourceKubernetes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	state := Kubernetes{
		Name: pkg.FromStr(k8sCluster.Name),
	}

	for state := range WaitForKubernetes(ctx, r.Client(), r.Organization(), k8sCluster.ID, 1*time.Second) {
		tflog.Info(ctx, "cluster state changed", map[string]any{
			"state": state.Status,
		})
		if state.Status == "FAILED" {
			resp.Diagnostics.AddError("failed to provision kubernetes cluster", state.Status)
		}
	}

	// Get kubeconfig
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.Client(), r.Organization(), k8sCluster.ID)
	if kubeConfigRes.HasError() {
		resp.Diagnostics.AddWarning("failed to get kubeconfig", kubeConfigRes.Error().Error())
	} else {
		kubeconfig := kubeConfigRes.Payload()
		state.KubeConfig = pkg.FromStr(string(*kubeconfig))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	fmt.Printf("\nCREATED: %s\n\n", k8sCluster.ID)
}

func (r *ResourceKubernetes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	identity := helper.From[KubernetesIdentity](ctx, req.Identity, &resp.Diagnostics)
	state := Kubernetes{}

	kubernetesRes := tmp.GetKubernetes(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("Failed to get kubernetes instance", kubernetesRes.Error().Error())
	}

	k8sInfo := kubernetesRes.Payload()
	state.Name = pkg.FromStr(k8sInfo.Name)

	// Get kubeconfig
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.Client(), r.Organization(), identity.ID.ValueString())
	if !kubeConfigRes.HasError() {
		kubeconfig := kubeConfigRes.Payload()
		state.KubeConfig = pkg.FromStr(string(*kubeconfig))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceKubernetes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	/*plan := helper.PlanFrom[Kubernetes](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}*/
	// TODO

	state := helper.From[Kubernetes](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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

// WaitForKubernetes surveille le status d'un cluster Kubernetes et renvoie un canal
// qui émet le cluster au premier appel et à chaque changement du champ Status.
// Le canal est fermé automatiquement quand le status devient ACTIVE ou FAILED (états terminaux).
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
