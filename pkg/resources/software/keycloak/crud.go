package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.dev/sdk/models"
)

// Create a new resource
func (r *ResourceKeycloak) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceKeycloak.Create()")

	plan := helper.PlanFrom[Keycloak](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	addonID, createDiags := addon.Create(ctx, r, &plan)
	res.Diagnostics.Append(createDiags...)
	if res.Diagnostics.HasError() {
		return
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}

	addon.SyncNetworkGroups(ctx, r, addonID, plan.Networkgroups, &plan.Networkgroups, &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceKeycloak) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceKeycloak.Read()")

	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonIsDeleted, diags := addon.Read(ctx, r, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if addonIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceKeycloak) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceKeycloak.Update()")

	plan := helper.PlanFrom[Keycloak](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonID, updateDiags := addon.Update(ctx, r, &plan, &state)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save target version before SetFromResponse overwrites it with current API value
	targetVersion := plan.Version

	plan.SetFromResponse(ctx, r.Client(), r.Organization(), addonID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle version upgrade (compare original plan version against state)
	if !targetVersion.IsNull() && !targetVersion.IsUnknown() && !targetVersion.Equal(state.Version) {
		tflog.Debug(ctx, "need version upgrade", map[string]any{
			"current_version": state.Version.ValueString(),
			"target_version":  targetVersion.ValueString(),
		})

		versionRes := r.
			SDK.
			V4().
			AddonProviders().
			AddonKeycloak().
			Addons().
			Addonkeycloakid(plan.ID.ValueString()).
			Version().
			Update().
			Createversionupdatekeycloak(ctx, &models.KeycloakPatchRequest{
				TargetVersion: targetVersion.ValueString(),
			})
		if versionRes.HasError() {
			resp.Diagnostics.AddError("failed to update Keycloak version", versionRes.Error().Error())
			return
		} else {
			kc := versionRes.Payload()
			plan.Version = pkg.FromStr(kc.Version)
			plan.Host = pkg.FromStr(kc.AccessURL)
			plan.AdminUsername = pkg.FromStr(kc.InitialCredentials.User)
			plan.AdminPassword = pkg.FromStr(kc.InitialCredentials.Password)
			plan.AccessDomain = pkg.FromStr(kc.EnvVars["CC_KEYCLOAK_HOSTNAME"])
			plan.FSBucketID = types.StringPointerValue(kc.Resources.FsbucketID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addon.SyncNetworkGroups(ctx, r, addonID, plan.Networkgroups, &plan.Networkgroups, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourceKeycloak) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourceKeycloak.Delete()")

	state := helper.StateFrom[Keycloak](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(addon.Delete(ctx, r, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
