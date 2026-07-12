package metabase

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// ModifyPlan recomputes the effective `tags_all` (provider defaults merged with the
// resource `tags`) at plan time, so a provider-level default_tags change propagates to
// existing resources.
func (r *ResourceMetabase) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.Provider == nil {
		return
	}

	plan := helper.PlanFrom[Metabase](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.TagsAll = pkg.ComputeTagsAll(ctx, r.DefaultTags(), plan.Tags, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

// Create a new resource
func (r *ResourceMetabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	createdAddon := res.Payload()

	mb.ID = pkg.FromStr(createdAddon.RealID)
	mb.Region = pkg.FromStr(createdAddon.Region)
	resp.Diagnostics.Append(resp.State.Set(ctx, mb)...)

	metabaseRes := tmp.GetMetabase(ctx, r.Client(), createdAddon.RealID)
	if metabaseRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase", metabaseRes.Error().Error())
	} else {
		metabase := metabaseRes.Payload()
		mb.Host = pkg.FromStr(metabase.AccessURL)
	}

	resources.SyncTags(ctx, r, resources.AddonTags, createdAddon.ID, mb.Tags, &mb.Tags, &mb.TagsAll, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, mb)...)
}

// Read resource information
func (r *ResourceMetabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	addonMBRes := tmp.GetMetabase(ctx, r.Client(), state.ID.ValueString())
	if addonMBRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if addonMBRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase resource", addonMBRes.Error().Error())
	} else {
		metabase := addonMBRes.Payload()
		state.Name = pkg.FromStr(metabase.Name)
		state.Host = pkg.FromStr(metabase.AccessURL)
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonId)
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase addon", addonRes.Error().Error())
	} else {
		addonPayload := addonRes.Payload()
		state.Region = pkg.FromStr(addonPayload.Region)
	}

	state.Tags, state.TagsAll = resources.ReadTags(ctx, r, resources.AddonTags, addonId, state.Tags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceMetabase) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	} else {
		state.Name = pkg.FromStr(addonRes.Payload().Name)
	}

	if addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
	} else {
		resources.SyncTags(ctx, r, resources.AddonTags, addonID, plan.Tags, &state.Tags, &state.TagsAll, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceMetabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	mb := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Metabase DELETE", map[string]any{"mb": mb})

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
