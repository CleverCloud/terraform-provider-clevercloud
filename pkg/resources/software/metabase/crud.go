package metabase

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceMetabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceMetabase.Create()")

	mb := helper.PlanFrom[Metabase](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "metabase")
	plan := prov.FirstPlan()
	if plan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       mb.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "metabase",
		Region:     mb.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	addon := res.Payload()

	mb.ID = pkg.FromStr(addon.RealID)
	mb.Region = pkg.FromStr(addon.Region)
	resp.Diagnostics.Append(resp.State.Set(ctx, mb)...)

	metabaseRes := tmp.GetMetabase(ctx, r.Client(), addon.RealID)
	if metabaseRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase", metabaseRes.Error().Error())
	} else {
		metabase := metabaseRes.Payload()
		mb.Host = pkg.FromStr(metabase.AccessURL)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, mb)...)
}

// Read resource information
func (r *ResourceMetabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceMetabase.Read()")

	state := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonId)
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase addon", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()
	state.Name = pkg.FromStr(addon.Name)
	state.Region = pkg.FromStr(addon.Region)

	addonMBRes := tmp.GetMetabase(ctx, r.Client(), state.ID.ValueString())
	if addonMBRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonMBRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase resource", addonMBRes.Error().Error())
		return
	}
	metabase := addonMBRes.Payload()
	state.Host = pkg.FromStr(metabase.AccessURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceMetabase) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceMetabase.Update()")

	plan := helper.PlanFrom[Metabase](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("metabase cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Metabase", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceMetabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	mb := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ResourceMetabase.Delete()")

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), mb.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonId)
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
