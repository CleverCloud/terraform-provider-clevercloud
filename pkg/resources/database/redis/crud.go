package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceRedis) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	rd := helper.PlanFrom[Redis](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	provider := pkg.LookupAddonProvider(*addonsProviders, "redis-addon")
	plan := pkg.LookupProviderPlan(provider, rd.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("This plan does not exists", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       rd.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "redis-addon",
		Region:     rd.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	rd.ID = pkg.FromStr(res.Payload().RealID)
	rd.CreationDate = pkg.FromI(res.Payload().CreationDate)

	resp.Diagnostics.Append(resp.State.Set(ctx, rd)...)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), rd.ID.ValueString())
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	rd.FromEnv(ctx, *envRes.Payload(), &resp.Diagnostics)

	addon.SyncNetworkGroups(
		ctx,
		r,
		res.Payload().ID,
		rd.Networkgroups,
		&rd.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, rd)...)
}

// Read resource information
func (r *ResourceRedis) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Redis READ", map[string]any{"request": req})

	var rd Redis
	diags := req.State.Get(ctx, &rd)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if rd.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), rd.ID.ValueString())
	if addonRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis resource", addonRes.Error().Error())
	}

	addonRD := addonRes.Payload()
	tflog.Debug(ctx, "redis", map[string]any{"payload": fmt.Sprintf("%+v", addonRD)})

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), rd.ID.ValueString())
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	rd.FromAddon(ctx, addonRD, &resp.Diagnostics)
	rd.FromEnv(ctx, *envRes.Payload(), &resp.Diagnostics)

	rd.Networkgroups = resources.ReadNetworkGroups(ctx, r, rd.ID.ValueString(), &resp.Diagnostics)

	diags = resp.State.Set(ctx, rd)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceRedis) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Redis](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Redis](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("redis cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Redis", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	addon.SyncNetworkGroups(
		ctx,
		r,
		plan.ID.ValueString(),
		plan.Networkgroups,
		&state.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceRedis) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	rd := Redis{}

	diags := req.State.Get(ctx, &rd)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Redis DELETE", map[string]any{"rd": rd})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), rd.ID.ValueString())
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
