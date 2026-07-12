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
)

// ModifyPlan recomputes the effective `tags_all` (provider defaults merged with the
// resource `tags`) at plan time, so a provider-level default_tags change propagates to
// existing resources.
func (r *ResourceAddon) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.Provider == nil {
		return
	}

	plan := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.TagsAll = pkg.ComputeTagsAll(ctx, r.DefaultTags(), plan.Tags, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

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

	addonReq := tmp.AddonRequest{
		Name:       ad.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: provider.ID,
		Region:     ad.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}

	ad.ID = pkg.FromStr(res.Payload().ID)
	ad.CreationDate = pkg.FromI(res.Payload().CreationDate)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), res.Payload().ID)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", res.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*envRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	resources.SyncTags(ctx, r, resources.AddonTags, res.Payload().ID, ad.Tags, &ad.Tags, &ad.TagsAll, &resp.Diagnostics)

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

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonRes.IsNotFoundError() {
		req.State.RemoveResource(ctx)
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

	ad.Tags, ad.TagsAll = resources.ReadTags(ctx, r, resources.AddonTags, ad.ID.ValueString(), ad.Tags, &resp.Diagnostics)

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
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Addon", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resources.SyncTags(ctx, r, resources.AddonTags, addonRes.Payload().ID, plan.Tags, &state.Tags, &state.TagsAll, &resp.Diagnostics)

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
