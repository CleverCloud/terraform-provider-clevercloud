package keycloak

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

const (
	envVarKeycloakRealms = "CC_KEYCLOAK_REALMS"
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
	// Handle realms option
	if realmsCommaSeparated := plan.GetRealmsCommaSeparated(ctx); realmsCommaSeparated != "" {
		addonReq.Options["realms"] = realmsCommaSeparated
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
		plan.Name = pkg.FromStr(keycloak.Name)
		plan.Host = pkg.FromStr(keycloak.AccessURL)
		plan.AdminUsername = pkg.FromStr(keycloak.InitialCredentials.User)
		plan.AdminPassword = pkg.FromStr(keycloak.InitialCredentials.Password)
		plan.Version = pkg.FromStr(keycloak.Version)
		plan.AccessDomain = pkg.FromStr(keycloak.EnvVars["CC_KEYCLOAK_HOSTNAME"])
		plan.FSBucketID = types.StringPointerValue(keycloak.Resources.FsbucketID)

		// Read realms from env vars if available
		if realmsValue, ok := keycloak.EnvVars[envVarKeycloakRealms]; ok && realmsValue != "" {
			realmsList := strings.Split(realmsValue, ",")
			plan.SetRealms(ctx, realmsList, &res.Diagnostics)
		}
		// If envVarKeycloakRealms not present, keep plan value (API doesn't confirm creation)
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceKeycloak) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.From[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Keycloak addon", addonRes.Error().Error())
	} else {
		addon := addonRes.Payload()
		state.Name = pkg.FromStr(addon.Name)
		state.Region = pkg.FromStr(addon.Region)
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

		// Read realms from env vars if available
		if realmsValue, ok := keycloak.EnvVars[envVarKeycloakRealms]; ok && realmsValue != "" {
			realmsList := strings.Split(realmsValue, ",")
			state.SetRealms(ctx, realmsList, &resp.Diagnostics)
		}
		// If envVarKeycloakRealms is not present, keep existing state value
		// This is necessary because we can't detect which realms actually exist
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
		state.Name = pkg.FromStr(addonRes.Payload().Name)
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
			state.AccessDomain = pkg.FromStr(kc.EnvVars["CC_KEYCLOAK_HOSTNAME"])
			state.FSBucketID = types.StringPointerValue(kc.Resources.FsbucketID)
		}
	}

	// Handle realms update
	planRealms := plan.GetRealms(ctx)
	stateRealms := state.GetRealms(ctx)

	if !areRealmsEqual(planRealms, stateRealms) {
		tflog.Debug(ctx, "realms changed, updating via env var", map[string]any{
			"old_realms": stateRealms,
			"new_realms": planRealms,
		})

		// Get the latest keycloak info to ensure we have entrypoint ID
		keycloakRes := r.SDK.
			V4().
			Keycloaks().
			Organisations().
			Ownerid(r.Organization()).
			Keycloaks().
			Addonkeycloakid(state.ID.ValueString()).
			Getkeycloak(ctx)

		if keycloakRes.HasError() {
			resp.Diagnostics.AddError("failed to get keycloak for realm update", keycloakRes.Error().Error())
			return
		}

		keycloak := keycloakRes.Payload()
		entrypointAppID := keycloak.Resources.Entrypoint

		if entrypointAppID == "" {
			resp.Diagnostics.AddError("missing entrypoint app", "cannot update realms without entrypoint app ID")
			return
		}

		// Get current env vars from the entrypoint app
		envRes := tmp.GetAppEnv(ctx, r.Client(), r.Organization(), entrypointAppID)
		if envRes.HasError() {
			resp.Diagnostics.AddError("failed to get app env vars", envRes.Error().Error())
			return
		}

		// Convert env vars to map using pkg.Reduce
		currentEnvs := pkg.Reduce(*envRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
			acc[e.Name] = e.Value
			return acc
		})

		// Update realms environment variable
		currentEnvs[envVarKeycloakRealms] = strings.Join(planRealms, ",")

		// Apply updated env vars
		updateEnvRes := tmp.UpdateAppEnv(ctx, r.Client(), r.Organization(), entrypointAppID, currentEnvs)
		if updateEnvRes.HasError() {
			resp.Diagnostics.AddError("failed to update app env vars", updateEnvRes.Error().Error())
			return
		}

		// Restart the entrypoint app to apply changes
		restartRes := tmp.RestartApp(ctx, r.Client(), r.Organization(), entrypointAppID)
		if restartRes.HasError() {
			// Error 4014 = app never deployed, can be ignored
			if apiErr, ok := restartRes.Error().(*client.APIError); !ok || apiErr.Code != "4014" {
				resp.Diagnostics.AddError("failed to restart app", restartRes.Error().Error())
				return
			}
		}

		state.SetRealms(ctx, planRealms, &resp.Diagnostics)
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

// areRealmsEqual compares two realm slices for equality (order-independent)
func areRealmsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aSet := make(map[string]bool, len(a))
	for _, realm := range a {
		aSet[realm] = true
	}

	for _, realm := range b {
		if !aSet[realm] {
			return false
		}
	}

	return true
}
