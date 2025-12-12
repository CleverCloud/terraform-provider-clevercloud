package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

// Create a new resource
func (r *ResourceKeycloak) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	plan := helper.From[Keycloak](ctx, req.Plan, &res.Diagnostics)
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
		Options:    map[string]string{},
	}

	if !plan.AccessDomain.IsNull() && !plan.AccessDomain.IsUnknown() {
		addonReq.Options["access-domain"] = plan.AccessDomain.ValueString()
	}
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		addonReq.Options["version"] = plan.Version.ValueString()
	}

	createAddonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Keycloak", createAddonRes.Error().Error())
		return
	}
	addon := createAddonRes.Payload()

	plan.ID = pkg.FromStr(addon.RealID)
	plan.Region = pkg.FromStr(addon.Region)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)

	keycloakRes := r.SDK.
		V4().
		Keycloaks().
		Organisations().
		Ownerid(r.Organization()).
		Keycloaks().
		Addonkeycloakid(addon.RealID).
		Getkeycloak(ctx)
	if keycloakRes.HasError() {
		res.Diagnostics.AddError("failed to get Keycloak", keycloakRes.Error().Error())
	} else {
		keycloak := keycloakRes.Payload()
		plan.Host = pkg.FromStr(keycloak.AccessURL)
		plan.AdminUsername = pkg.FromStr(keycloak.InitialCredentials.User)
		plan.AdminPassword = pkg.FromStr(keycloak.InitialCredentials.Password)
		plan.Version = pkg.FromStr(keycloak.Version)
		plan.AccessDomain = pkg.FromStr(keycloak.EnvVars["CC_KEYCLOAK_HOSTNAME"])
		plan.FSBucketID = types.StringPointerValue(keycloak.Resources.FsbucketID)
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceKeycloak) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.From[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	keycloakRes := r.SDK.
		V4().
		Keycloaks().
		Organisations().
		Ownerid(r.Organization()).
		Keycloaks().
		Addonkeycloakid(state.ID.ValueString()).
		Getkeycloak(ctx)
	if keycloakRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if keycloakRes.HasError() {
		resp.Diagnostics.AddError("failed to get keycloak", keycloakRes.Error().Error())
	} else {
		keycloak := keycloakRes.Payload()
		state.Host = pkg.FromStr(keycloak.AccessURL)
		state.AdminUsername = pkg.FromStr(keycloak.InitialCredentials.User)
		state.AdminPassword = pkg.FromStr(keycloak.InitialCredentials.Password)
		state.Version = pkg.FromStr(keycloak.Version)
		state.AccessDomain = pkg.FromStr(keycloak.EnvVars["CC_KEYCLOAK_HOSTNAME"])
		state.FSBucketID = types.StringPointerValue(keycloak.Resources.FsbucketID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceKeycloak) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.From[Keycloak](ctx, req.Plan, &resp.Diagnostics)
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
	} else {
		state.Name = plan.Name
	}

	// Handle version upgrade
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() && !plan.Version.Equal(state.Version) {
		tflog.Debug(ctx, "need version upgrade", map[string]any{
			"current_version": state.Version.ValueString(),
			"target_version":  plan.Version.ValueString(),
		})

		versionRes := r.
			SDK.
			V4().
			AddonProviders().
			AddonKeycloak().
			Addons().
			Addonkeycloakid(state.ID.ValueString()).
			Version().
			Update().
			Createversionupdatekeycloak(ctx, &models.KeycloakPatchRequest{
				TargetVersion: plan.Version.ValueString(),
			})
		if versionRes.HasError() {
			resp.Diagnostics.AddError("failed to update Keycloak version", versionRes.Error().Error())
			return
		} else {
			kc := versionRes.Payload()
			state.Version = pkg.FromStr(kc.Version)
			state.Host = pkg.FromStr(kc.AccessURL)
			state.AdminUsername = pkg.FromStr(kc.InitialCredentials.User)
			state.AdminPassword = pkg.FromStr(kc.InitialCredentials.Password)
			state.Version = pkg.FromStr(kc.Version)
			state.AccessDomain = pkg.FromStr(kc.EnvVars["CC_KEYCLOAK_HOSTNAME"])

		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceKeycloak) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.From[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete addon", res.Error().Error())
	} else {
		resp.State.RemoveResource(ctx)
	}
}
