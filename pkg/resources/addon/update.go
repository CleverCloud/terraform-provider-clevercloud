package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceAddon) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("addon cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Addon", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update centralizes the common Update logic for all addon resources.
// Uses the plan pattern: writes the plan to state after updating, not the old state.
// Returns the addon_xxx ID (needed for SyncNetworkGroups) and diagnostics.
func Update[T AddonPlan](ctx context.Context, r AddonResource, plan, state T) (addonID string, diags diag.Diagnostics) {
	planCommon := plan.GetCommonPtr()
	stateCommon := state.GetCommonPtr()

	tflog.Debug(ctx, "addon.Update()", map[string]any{"provider": r.GetProviderSlug()})

	if planCommon.ID.ValueString() != stateCommon.ID.ValueString() {
		diags.AddError("addon cannot be updated", "mismatched IDs")
		return "", diags
	}

	// Convert RealID to AddonID
	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), planCommon.ID.ValueString())
	if err != nil {
		diags.AddError("failed to get addon ID", err.Error())
		return "", diags
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), addonID, map[string]string{
		"name": planCommon.Name.ValueString(),
	})
	if addonRes.HasError() {
		diags.AddError("failed to update addon", addonRes.Error().Error())
		return addonID, diags
	}

	// Refresh common fields from API response to fill Computed values
	a := addonRes.Payload()
	planCommon.ID = pkg.FromStr(a.RealID)
	planCommon.Name = pkg.FromStr(a.Name)
	planCommon.Plan = pkg.FromStr(a.Plan.Slug)
	planCommon.Region = pkg.FromStr(a.Region)
	planCommon.CreationDate = pkg.FromI(a.CreationDate)

	// Callers must call plan.SetFromResponse() after any custom operations
	// to fill provider-specific Computed fields (host, port, password...).

	return addonID, diags
}
