package generic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	helperAddon "go.clever-cloud.com/terraform-provider/pkg/helper/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceAddon) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceAddon.Create()")
	ad := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan, diags := helperAddon.GetAddonPlan(ctx, r.Client(), ad.ThirdPartyProvider.ValueString(), ad.Plan.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       ad.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: ad.ThirdPartyProvider.ValueString(),
		Region:     ad.Region.ValueString(),
	}

	createdAd, diags := helperAddon.Create(ctx, r.Client(), r.Organization(), addonReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ad.ID = pkg.FromStr(createdAd.ID)
	ad.Name = pkg.FromStr(createdAd.Name)
	ad.Region = pkg.FromStr(createdAd.Region)
	ad.CreationDate = pkg.FromI(createdAd.CreationDate)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), createdAd.ID)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", envRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*envRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
}

// Read resource information
func (r *ResourceAddon) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceAddon.Read()")

	ad := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Objects
	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on", addonRes.Error().Error())
		return
	}

	addon := addonRes.Payload()

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*addonEnvRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})

	ad.Name = pkg.FromStr(addon.Name)
	ad.Plan = pkg.FromStr(addon.Plan.Slug)
	ad.Region = pkg.FromStr(addon.Region)
	ad.ThirdPartyProvider = pkg.FromStr(addon.Provider.ID)
	ad.CreationDate = pkg.FromI(addon.CreationDate)
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
}

// Update resource
func (r *ResourceAddon) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceAddon.Update()")
	plan := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("addon cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	diags := helperAddon.Update(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceAddon) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourceAddon.Delete()")
	ad := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := helperAddon.Delete(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
