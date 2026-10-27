package matomo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/addonid"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/addons"
)

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

	plan := pkg.FirstPlan(provider)
	if plan == nil {
		res.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := models.WannabeAddonProvision{
		Name:       appMatomo.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "addon-matomo",
		Region:     appMatomo.Region.ValueString(),
	}

	createAddonRes := resources.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Matomo", createAddonRes.Error().Error())
		return
	}
	addon := createAddonRes.Payload()

	appMatomo.ID = pkg.FromStr(pkg.Deref(addon.RealID))
	appMatomo.Region = pkg.FromStr(addon.Region)
	res.Diagnostics.Append(res.State.Set(ctx, appMatomo)...)

	matomoRes := r.SDK.V4().AddonProviders().AddonMatomo().Addons().Addonmatomoid(pkg.Deref(addon.RealID)).Getmatomo(ctx)
	if matomoRes.HasError() {
		res.Diagnostics.AddError("cannot get matomo", matomoRes.Error().Error())
	} else {
		matomo := matomoRes.Payload()
		appMatomo.Host = pkg.FromStr(matomo.AccessURL)
		appMatomo.Version = pkg.FromStr(matomo.Version)
	}

	appMatomo.Tags = resources.SyncAddonTags(ctx, r, addon.ID, appMatomo.Tags, resources.NullTags, &res.Diagnostics)

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

	matomoRes := r.SDK.V4().AddonProviders().AddonMatomo().Addons().Addonmatomoid(state.ID.ValueString()).Getmatomo(ctx)
	if matomoRes.HasError() {
		resp.Diagnostics.AddError("cannot get matomo", matomoRes.Error().Error())
	} else {
		matomo := matomoRes.Payload()
		state.Name = pkg.FromStr(matomo.Name)
		state.Host = pkg.FromStr(matomo.AccessURL)
		state.Version = pkg.FromStr(matomo.Version)
	}

	addonId, err := addonid.FromRealID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(addonId).Getaddonbyorgaandaddonid(ctx)
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Matomo addon", addonRes.Error().Error())
	} else {
		addon := addonRes.Payload()
		state.Region = pkg.FromStr(addon.Region)
	}

	state.Tags = resources.ReadAddonTags(ctx, r, addonId, state.Tags, &resp.Diagnostics)

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
	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(plan.ID.ValueString()).Updateaddoninfo(ctx, addons.UpdateaddoninfoBody{
		Name: plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Matomo", addonRes.Error().Error())
	} else {
		state.Name = pkg.FromStr(pkg.Deref(addonRes.Payload().Name))
	}

	if !plan.Tags.Equal(state.Tags) {
		state.Tags = resources.SyncAddonTags(ctx, r, state.ID.ValueString(), plan.Tags, state.Tags, &resp.Diagnostics)
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

	addonID, err := addonid.FromRealID(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("cannot get ID of matomo", err.Error())
	}

	res := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(addonID).Deprovisionaddonbyorgaandaddonid(ctx)
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
