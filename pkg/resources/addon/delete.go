package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceAddon) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ad := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete add-on", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// Delete centralizes the common Delete logic for all addon resources
func Delete[T AddonPlan](ctx context.Context, r AddonResource, state T) diag.Diagnostics {
	diags := diag.Diagnostics{}
	common := state.GetCommonPtr()

	tflog.Debug(ctx, "addon.Delete()", map[string]any{"provider": r.GetSlug()})

	// Convert RealID to AddonID
	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), common.ID.ValueString())
	if err != nil {
		diags.AddError("failed to get addon ID", err.Error())
		return diags
	}

	// Delete the addon
	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonID)
	if res.IsNotFoundError() {
		return diags
	}
	if res.HasError() {
		diags.AddError("failed to delete addon", res.Error().Error())
		return diags
	}

	return diags
}
