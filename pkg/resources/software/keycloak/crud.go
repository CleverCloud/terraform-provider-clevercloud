package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceKeycloak) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	plan := helper.PlanFrom[Keycloak](ctx, req.Plan, &res.Diagnostics)
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

	addonPlan := provider.FirstPlan()
	if addonPlan == nil {
		res.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       plan.Name.ValueString(),
		Plan:       addonPlan.ID,
		ProviderID: "keycloak",
		Region:     plan.Region.ValueString(),
	}

	createAddonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Keycloak", createAddonRes.Error().Error())
		return
	}

	plan.ID = pkg.FromStr(createAddonRes.Payload().RealID)
	res.Diagnostics.Append(res.State.Set(ctx, plan)...)

	kcEnvRes := tmp.GetKeycloak(ctx, r.Client(), createAddonRes.Payload().RealID)
	if kcEnvRes.HasError() {
		res.Diagnostics.AddError("failed to get Keycloak connection infos", kcEnvRes.Error().Error())
		return
	}

	kc := kcEnvRes.Payload()
	plan.Host = pkg.FromStr(kc.AccessURL)
	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceKeycloak) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	kcRes := tmp.GetKeycloak(ctx, r.Client(), state.ID.ValueString())
	if kcRes.HasError() {
		resp.Diagnostics.AddError("cannot get Keycloak", kcRes.Error().Error())
	}

	state.Host = pkg.FromStr(kcRes.Payload().AccessURL)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceKeycloak) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Keycloak](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
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
}

// Delete resource
func (r *ResourceKeycloak) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
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
