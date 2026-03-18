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
	"go.clever-cloud.dev/client"
)

// ReadRes represents the response from reading an addon
type ReadRes struct {
	Addon          tmp.AddonResponse
	AddonID        string // addon_xxx format
	AddonIsDeleted bool
}

func (r *ReadRes) GetAddon() *tmp.AddonResponse {
	return &r.Addon
}

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

// ReadAddon handles the low-level API calls for reading an addon.
// Handles legacy addon_xxx IDs transparently (converts to RealID).
func ReadAddon(ctx context.Context, cc *client.Client, org, id string) (*ReadRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	r := &ReadRes{}

	// Guard: empty ID means resource was never fully created
	if id == "" {
		r.AddonIsDeleted = true
		return r, diags
	}

	// Handle legacy addon_xxx IDs from imports
	realID, err := tmp.AddonIDToRealID(ctx, cc, org, id)
	if err != nil {
		diags.AddError("failed to resolve addon ID", err.Error())
		return r, diags
	}

	// Convert RealID to AddonID for API calls
	addonID, err := tmp.RealIDToAddonID(ctx, cc, org, realID)
	if err != nil {
		diags.AddError("failed to get addon ID", err.Error())
		return r, diags
	}

	// Fetch addon metadata
	addonRes := tmp.GetAddon(ctx, cc, org, addonID)
	if addonRes.IsNotFoundError() {
		r.AddonIsDeleted = true
		return r, diags
	}
	if addonRes.HasError() {
		diags.AddError("failed to get addon", addonRes.Error().Error())
		return r, diags
	}

	r.Addon = *addonRes.Payload()
	r.AddonID = addonID

	return r, diags
}

// Read centralizes the common Read logic for all addon resources.
// Returns true if the addon is deleted and should be removed from state.
func Read[T AddonPlan](ctx context.Context, r AddonResource, state T) (addonIsDeleted bool, diags diag.Diagnostics) {
	common := state.GetCommonPtr()

	tflog.Debug(ctx, "addon.Read()", map[string]any{"provider": r.GetSlug()})

	// Call common ReadAddon function
	readRes, readDiags := ReadAddon(ctx, r.Client(), r.Organization(), common.ID.ValueString())
	diags.Append(readDiags...)
	if diags.HasError() {
		return false, diags
	}

	// Check if addon was deleted
	if readRes.AddonIsDeleted {
		return true, diags
	}

	// Map common fields from API response
	a := readRes.GetAddon()
	common.ID = pkg.FromStr(a.RealID)
	common.Name = pkg.FromStr(a.Name)
	common.Region = pkg.FromStr(a.Region)
	common.Plan = pkg.FromStr(a.Plan.Slug)
	common.CreationDate = pkg.FromI(a.CreationDate)

	// Read provider-specific fields (host, port, password...)
	state.SetFromResponse(ctx, r.Client(), r.Organization(), readRes.AddonID, &diags)
	if diags.HasError() {
		return false, diags
	}

	// Read networkgroups
	common.Networkgroups = resources.ReadNetworkGroups(ctx, r, readRes.AddonID, &diags)

	return false, diags
}
