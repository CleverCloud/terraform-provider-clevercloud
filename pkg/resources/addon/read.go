package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceAddon) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ad := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on", addonRes.Error().Error())
		return
	}

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*addonEnvRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})

	a := addonRes.Payload()
	ad.Name = pkg.FromStr(a.Name)
	ad.Plan = pkg.FromStr(a.Plan.Slug)
	ad.Region = pkg.FromStr(a.Region)
	ad.ThirdPartyProvider = pkg.FromStr(a.Provider.ID)
	ad.CreationDate = pkg.FromI(a.CreationDate)
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
}

// Read centralizes the common Read logic for all addon resources.
// Returns true if the addon is deleted and should be removed from state.
func Read[T AddonPlan](ctx context.Context, r AddonResource, state T) (addonIsDeleted bool, diags diag.Diagnostics) {
	common := state.GetCommonPtr()

	tflog.Debug(ctx, "addon.Read()", map[string]any{"provider": r.GetProviderSlug()})

	// Guard: empty ID means resource was never fully created
	if common.ID.ValueString() == "" {
		return true, diags
	}

	// Convert RealID to AddonID
	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), common.ID.ValueString())
	if err != nil {
		diags.AddError("failed to get addon ID", err.Error())
		return false, diags
	}

	// Fetch addon metadata
	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonID)
	if addonRes.IsNotFoundError() {
		return true, diags
	}
	if addonRes.HasError() {
		diags.AddError("failed to get addon", addonRes.Error().Error())
		return false, diags
	}

	// Map common fields from API response
	a := addonRes.Payload()
	common.ID = pkg.FromStr(a.RealID)
	common.Name = pkg.FromStr(a.Name)
	common.Region = pkg.FromStr(a.Region)
	common.Plan = pkg.FromStr(a.Plan.Slug)
	common.CreationDate = pkg.FromI(a.CreationDate)

	// Read provider-specific fields (host, port, password...)
	state.SetFromResponse(ctx, r.Client(), r.Organization(), addonID, &diags)
	if diags.HasError() {
		return false, diags
	}

	// Read networkgroups
	common.Networkgroups = resources.ReadNetworkGroups(ctx, r, addonID, &diags)

	return false, diags
}
