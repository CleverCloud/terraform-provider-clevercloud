package matomo

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
func (r *ResourceMatomo) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.Provider == nil {
		return
	}

	plan := helper.PlanFrom[Matomo](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.TagsAll = pkg.ComputeTagsAll(ctx, r.DefaultTags(), plan.Tags, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

// Create a new resource
func (r *ResourceMatomo) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	appMatomo := helper.PlanFrom[Matomo](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		res.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	provider := pkg.LookupAddonProvider(*addonsProviders, "addon-matomo")

	plan := provider.FirstPlan()
	if plan == nil {
		res.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       appMatomo.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "addon-matomo",
		Region:     appMatomo.Region.ValueString(),
	}

	createAddonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Matomo", createAddonRes.Error().Error())
		return
	}
	createdAddon := createAddonRes.Payload()

	appMatomo.ID = pkg.FromStr(createdAddon.RealID)
	appMatomo.Region = pkg.FromStr(createdAddon.Region)
	res.Diagnostics.Append(res.State.Set(ctx, appMatomo)...)

	matomoRes := tmp.GetMatomo(ctx, r.Client(), createdAddon.RealID)
	if matomoRes.HasError() {
		res.Diagnostics.AddError("cannot get matomo", matomoRes.Error().Error())
	} else {
		matomo := matomoRes.Payload()
		appMatomo.Host = pkg.FromStr(matomo.AccessURL)
		appMatomo.Version = pkg.FromStr(matomo.Version)
	}

	resources.SyncTags(ctx, r, resources.AddonTags, createdAddon.ID, appMatomo.Tags, &appMatomo.Tags, &appMatomo.TagsAll, &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, appMatomo)...)
}

// Read resource information
func (r *ResourceMatomo) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[Matomo](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	matomoRes := tmp.GetMatomo(ctx, r.Client(), state.ID.ValueString())
	if matomoRes.HasError() {
		resp.Diagnostics.AddError("cannot get matomo", matomoRes.Error().Error())
	} else {
		matomo := matomoRes.Payload()
		state.Name = pkg.FromStr(matomo.Name)
		state.Host = pkg.FromStr(matomo.AccessURL)
		state.Version = pkg.FromStr(matomo.Version)
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonId)
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Matomo addon", addonRes.Error().Error())
	} else {
		addonPayload := addonRes.Payload()
		state.Region = pkg.FromStr(addonPayload.Region)
	}

	state.Tags, state.TagsAll = resources.ReadTags(ctx, r, resources.AddonTags, addonId, state.Tags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceMatomo) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Matomo](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Matomo](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("matomo cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Matomo", addonRes.Error().Error())
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
func (r *ResourceMatomo) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Matomo](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Matomo DELETE", map[string]any{"matomo": state})

	addonID, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("cannot get ID of matomo", err.Error())
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonID)
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
