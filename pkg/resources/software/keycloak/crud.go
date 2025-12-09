package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceKeycloak) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	kc := helper.PlanFrom[Keycloak](ctx, req.Plan, &res.Diagnostics)
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

	plan := provider.FirstPlan()
	if plan == nil {
		res.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
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
	addon := createAddonRes.Payload()

	kc.ID = pkg.FromStr(addon.RealID)
	res.Diagnostics.Append(res.State.Set(ctx, kc)...)

	keycloakRes := tmp.GetKeycloak(ctx, r.Client(), addon.RealID)
	if keycloakRes.HasError() {
		res.Diagnostics.AddError("failed to get Keycloak", keycloakRes.Error().Error())
	} else {
		keycloak := keycloakRes.Payload()
		kc.Host = pkg.FromStr(keycloak.AccessURL)
		kc.AdminUsername = pkg.FromStr(keycloak.InitialCredentials.AdminUsername)
		kc.AdminPassword = pkg.FromStr(keycloak.InitialCredentials.AdminPassword)
	}

	res.Diagnostics.Append(res.State.Set(ctx, kc)...)
}

// Read resource information
func (r *ResourceKeycloak) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Keycloak READ", map[string]any{"request": req})

	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	keycloakRes := tmp.GetKeycloak(ctx, r.Client(), state.ID.ValueString())
	if keycloakRes.HasError() {
		resp.Diagnostics.AddError("failed to get keycloak", keycloakRes.Error().Error())
	} else {
		keycloak := keycloakRes.Payload()
		state.Host = pkg.FromStr(keycloak.AccessURL)
		state.AdminUsername = pkg.FromStr(keycloak.InitialCredentials.AdminUsername)
		state.AdminPassword = pkg.FromStr(keycloak.InitialCredentials.AdminPassword)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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
	} else {
		state.Name = plan.Name
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceKeycloak) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	kc := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
