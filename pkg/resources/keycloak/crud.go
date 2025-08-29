package keycloak

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
func (r *ResourceKeycloak) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceKeycloak.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}

	tflog.Warn(ctx, "Keycloak product is still in beta, use it with care")
}

// Create a new resource
func (r *ResourceKeycloak) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	kc := helper.PlanFrom[Keycloak](ctx, req.Plan, &res.Diagnostics)
	res.Diagnostics.Append(req.Plan.Get(ctx, &kc)...)
	if res.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		res.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	provider := pkg.LookupAddonProvider(*addonsProviders, "keycloak")

	plan := pkg.LookupProviderPlan(provider, kc.Plan.ValueString())
	if plan == nil {
		res.Diagnostics.AddError("This plan does not exists", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       kc.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "keycloak",
		Region:     kc.Region.ValueString(),
	}

	createAddonRes := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Keycloak", createAddonRes.Error().Error())
		return
	}

	kc.ID = pkg.FromStr(createAddonRes.Payload().RealID)
	kc.CreationDate = pkg.FromI(createAddonRes.Payload().CreationDate)

	res.Diagnostics.Append(res.State.Set(ctx, kc)...)
	if res.Diagnostics.HasError() {
		return
	}

	kcEnvRes := tmp.GetAddonEnv(ctx, r.cc, r.org, kc.ID.ValueString())
	if kcEnvRes.HasError() {
		res.Diagnostics.AddError("failed to get Keycloak connection infos", kcEnvRes.Error().Error())
		return
	}

	kcEnv := *kcEnvRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", kcEnv),
	})

	hostEnvVar := pkg.First(kcEnv, func(v tmp.EnvVar) bool {
		return v.Name == "CC_KEYCLOAK_URL"
	})
	if hostEnvVar == nil {
		res.Diagnostics.AddError("cannot get Keycloak infos", "missing CC_KEYCLOAK_URL env var on created addon")
		return
	}

	kc.Host = pkg.FromStr(hostEnvVar.Value)

	res.Diagnostics.Append(res.State.Set(ctx, kc)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceKeycloak) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Keycloak READ", map[string]any{"request": req})

	var kc Keycloak
	diags := req.State.Get(ctx, &kc)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO

	diags = resp.State.Set(ctx, kc)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceKeycloak) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Keycloak](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("keycloak cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.cc, r.org, plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Keycloak", addonRes.Error().Error())
		return
	}
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourceKeycloak) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	kc := Keycloak{}

	diags := req.State.Get(ctx, &kc)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Keycloak DELETE", map[string]any{"keycloak": kc})

	res := tmp.DeleteAddon(ctx, r.cc, r.org, kc.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete addon", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// Import resource
func (r *ResourceKeycloak) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
