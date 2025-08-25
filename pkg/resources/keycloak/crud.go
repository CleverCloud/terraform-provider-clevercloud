package keycloak

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
func (r *ResourceKeycloak) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	kc := helper.PlanFrom[Keycloak](ctx, req.Plan, &res.Diagnostics)
	res.Diagnostics.Append(req.Plan.Get(ctx, &kc)...)
	if res.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
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

	createAddonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Keycloak", createAddonRes.Error().Error())
		return
	}

	kc.ID = pkg.FromStr(createAddonRes.Payload().RealID)
	kc.CreationDate = pkg.FromI(createAddonRes.Payload().CreationDate)

	res.Diagnostics.Append(res.State.Set(ctx, kc)...)

	kcEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), kc.ID.ValueString())
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
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
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

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), kc.ID.ValueString())
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
