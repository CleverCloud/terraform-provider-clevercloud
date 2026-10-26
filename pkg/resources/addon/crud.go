package addon

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/addons"
)

// Create a new resource
func (r *ResourceAddon) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ad := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	provider := pkg.LookupAddonProvider(*addonsProviders, ad.ThirdPartyProvider.ValueString())
	if provider == nil {
		resp.Diagnostics.AddError("This provider doesn't exist", fmt.Sprintf("available providers are: %s", strings.Join(pkg.AddonProvidersAsList(*addonsProviders), ", ")))
		return
	}

	plan := pkg.LookupProviderPlan(provider, ad.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("This plan doesn't exist", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := models.WannabeAddonProvision{
		Name:       ad.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: provider.ID,
		Region:     ad.Region.ValueString(),
	}

	res := resources.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}

	ad.ID = pkg.FromStr(res.Payload().ID)
	ad.CreationDate = pkg.FromI(pkg.Deref(res.Payload().CreationDate))

	envRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(res.Payload().ID).Env().Getaddonenvbyorgaandaddonid(ctx)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", envRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*envRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v models.AddonEnvironmentView) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	ad.Tags = resources.SyncAddonTags(ctx, r, ad.ID.ValueString(), ad.Tags, resources.NullTags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
}

// Read resource information
func (r *ResourceAddon) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Add-on READ", map[string]any{"request": req})

	var ad Addon
	diags := req.State.Get(ctx, &ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if ad.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(ad.ID.ValueString()).Getaddonbyorgaandaddonid(ctx)
	if addonRes.IsNotFoundError() {
		req.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on", addonRes.Error().Error())
		return
	}

	addonEnvRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(ad.ID.ValueString()).Env().Getaddonenvbyorgaandaddonid(ctx)
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*addonEnvRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v models.AddonEnvironmentView) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})

	a := addonRes.Payload()
	ad.Name = pkg.FromStr(pkg.Deref(a.Name))
	// Providers do not agree on the case of their plan slugs: jenkins answers S,
	// M, L while postgresql answers dev, xs_sml. LookupProviderPlan matches
	// case-insensitively, so both spellings create the same add-on; keep the one
	// already in state so a plan never drifts on case alone. An import has
	// nothing in state yet and takes the provider's spelling.
	plan := pkg.Deref(a.Plan.Slug)
	if configured := ad.Plan.ValueString(); strings.EqualFold(configured, plan) {
		plan = configured
	}
	ad.Plan = pkg.FromStr(plan)
	ad.Region = pkg.FromStr(a.Region)
	ad.ThirdPartyProvider = pkg.FromStr(a.Provider.ID)
	ad.CreationDate = pkg.FromI(pkg.Deref(a.CreationDate))
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	ad.Tags = resources.ReadAddonTags(ctx, r, ad.ID.ValueString(), ad.Tags, &resp.Diagnostics)

	diags = resp.State.Set(ctx, ad)
	resp.Diagnostics.Append(diags...)
}

// Update resource
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
	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(plan.ID.ValueString()).Updateaddoninfo(ctx, addons.UpdateaddoninfoBody{
		Name: plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Addon", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(pkg.Deref(addonRes.Payload().Name))

	if !plan.Tags.Equal(state.Tags) {
		state.Tags = resources.SyncAddonTags(ctx, r, state.ID.ValueString(), plan.Tags, state.Tags, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceAddon) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var ad Addon

	diags := req.State.Get(ctx, &ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Add-on DELETE", map[string]any{"addon": ad})

	res := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(ad.ID.ValueString()).Deprovisionaddonbyorgaandaddonid(ctx)
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
