package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceKubernetes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "ResourceKubernetes.Create()", map[string]any{"request": fmt.Sprintf("%+v", req.Plan)})

	plan := helper.PlanFrom[Kubernetes](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	prov := pkg.LookupAddonProvider(*addonsProviders, "kubernetes")
	addonPlan := pkg.LookupProviderPlan(prov, "beta")
	if addonPlan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       plan.Name.ValueString(),
		Plan:       addonPlan.ID,
		ProviderID: "kubernetes",
		Region:     plan.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create kubernetes addon", res.Error().Error())
		return
	}
	addon := res.Payload()

	tflog.Info(ctx, "CREATED", map[string]any{
		"k8s": fmt.Sprintf("%+v", addon),
	})

	state := Kubernetes{
		ID:     pkg.FromStr(addon.RealID),
		Name:   pkg.FromStr(addon.Name),
		Region: pkg.FromStr(addon.Region),
	}

	kubernetesRes := tmp.GetKubernetes(ctx, r.Client(), r.Organization(), addon.RealID)
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("failed to get kubernetes details", kubernetesRes.Error().Error())
	} else {
		k8s := kubernetesRes.Payload()
		fmt.Printf("%+v\n", k8s)

		state.Version = pkg.FromStr(k8s.Version)
	}

	// Get kubeconfig
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.Client(), r.Organization(), addon.RealID)
	if kubeConfigRes.HasError() {
		resp.Diagnostics.AddError("failed to get kubeconfig", kubeConfigRes.Error().Error())
	} else {
		kubeconfig := kubeConfigRes.Payload()
		state.KubeConfig = pkg.FromStr(string(*kubeconfig))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceKubernetes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "ResourceKubernetes.Read()", map[string]any{"request": req})

	state := helper.StateFrom[Kubernetes](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get kubernetes details
	kubernetesRes := tmp.GetKubernetes(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("Failed to get kubernetes instance", kubernetesRes.Error().Error())
		return
	} else {
		k8s := kubernetesRes.Payload()
		state.Version = pkg.FromStr(k8s.Version)
	}

	// Get kubeconfig
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if !kubeConfigRes.HasError() {
		kubeconfig := kubeConfigRes.Payload()
		state.KubeConfig = pkg.FromStr(string(*kubeconfig))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceKubernetes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "ResourceKubernetes.Update()", map[string]any{"request": fmt.Sprintf("%+v", req.Plan)})
}

// Delete resource
func (r *ResourceKubernetes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Kubernetes](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "ResourceKubernetes.Delete()", map[string]any{"kubernetes": state})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete kubernetes addon", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
