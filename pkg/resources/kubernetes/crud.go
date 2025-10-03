package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceKubernetes) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceKubernetes.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}
}

// Create a new resource
func (r *ResourceKubernetes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "ResourceKubernetes.Create()", map[string]any{"request": fmt.Sprintf("%+v", req.Plan)})

	plan := helper.PlanFrom[Kubernetes](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
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

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create kubernetes addon", res.Error().Error())
		return
	}
	addon := res.Payload()

	state := Kubernetes{
		ID:     pkg.FromStr(addon.RealID),
		Name:   pkg.FromStr(addon.Name),
		Region: pkg.FromStr(addon.Region),
	}

	kubernetesRes := tmp.GetKubernetes(ctx, r.cc, r.org, addon.RealID)
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("failed to get kubernetes details", kubernetesRes.Error().Error())
	} else {
		k8s := kubernetesRes.Payload()
		fmt.Printf("%+v\n", k8s)

		state.Version = pkg.FromStr(k8s.Version)
	}

	// Get kubeconfig using Raw type
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.cc, r.org, addon.RealID)
	if kubeConfigRes.HasError() {
		resp.Diagnostics.AddError("failed to get kubeconfig", kubeConfigRes.Error().Error())
	} else {
		raw := kubeConfigRes.Payload()
		fmt.Printf("%+v\n", raw)
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
	kubernetesRes := tmp.GetKubernetes(ctx, r.cc, r.org, state.ID.ValueString())
	if kubernetesRes.HasError() {
		resp.Diagnostics.AddError("Failed to get kubernetes instance", kubernetesRes.Error().Error())
		return
	} else {
		k8s := kubernetesRes.Payload()
		state.Version = pkg.FromStr(k8s.Version)
	}

	// Get kubeconfig using Raw type
	kubeConfigRes := tmp.GetKubeconfig(ctx, r.cc, r.org, state.ID.ValueString())
	if !kubeConfigRes.HasError() {
		raw := kubeConfigRes.Payload()
		// Convert []byte to string directly since client.Raw is []byte
		state.KubeConfig = pkg.FromStr(string(*raw))
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

	res := tmp.DeleteAddon(ctx, r.cc, r.org, state.ID.ValueString())
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

// Import resource
func (r *ResourceKubernetes) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
