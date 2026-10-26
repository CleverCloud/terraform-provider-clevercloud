package redis

import (
	"context"
	"fmt"
	"strconv"
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
	"go.clever-cloud.dev/sdk/app"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/addons"
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

	addonReq := models.WannabeAddonProvision{
		Name:       rd.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "redis-addon",
		Region:     rd.Region.ValueString(),
	}

	res := resources.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	rd.ID = pkg.FromStr(pkg.Deref(res.Payload().RealID))
	rd.CreationDate = pkg.FromI(pkg.Deref(res.Payload().CreationDate))

	resp.Diagnostics.Append(resp.State.Set(ctx, rd)...)

	envRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(rd.ID.ValueString()).Env().Getaddonenvbyorgaandaddonid(ctx)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	env := *envRes.Payload()
	envAsMap := pkg.Reduce(env, map[string]types.String{}, func(acc map[string]types.String, v models.AddonEnvironmentView) map[string]types.String {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", envAsMap),
	})
	port, err := strconv.ParseInt(envAsMap[app.REDIS_PORT].ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("invalid port received", "expect REDIS_PORT to be an Integer")
	}
	rd.Host = envAsMap[app.REDIS_HOST]
	rd.Port = pkg.FromI(port)
	rd.Token = envAsMap[app.REDIS_PASSWORD]

	addon.SyncNetworkGroups(
		ctx,
		r,
		res.Payload().ID,
		rd.Networkgroups,
		&rd.Networkgroups,
		&resp.Diagnostics,
	)
	rd.Tags = resources.SyncAddonTags(ctx, r, res.Payload().ID, rd.Tags, resources.NullTags, &resp.Diagnostics)

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

	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(rd.ID.ValueString()).Getaddonbyorgaandaddonid(ctx)
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

	envRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(rd.ID.ValueString()).Env().Getaddonenvbyorgaandaddonid(ctx)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	env := *envRes.Payload()
	envAsMap := pkg.Reduce(env, map[string]types.String{}, func(acc map[string]types.String, v models.AddonEnvironmentView) map[string]types.String {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", envAsMap),
	})
	port, err := strconv.ParseInt(envAsMap[app.REDIS_PORT].ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("invalid port received", "expect REDIS_PORT to be an Integer")
	}

	rd.Name = pkg.FromStr(pkg.Deref(addonRD.Name))
	rd.Host = envAsMap[app.REDIS_HOST]
	rd.Plan = pkg.FromStr(pkg.Deref(addonRD.Plan.Slug))
	rd.Port = pkg.FromI(port)
	rd.Region = pkg.FromStr(addonRD.Region)
	rd.Token = envAsMap[app.REDIS_PASSWORD]

	rd.Networkgroups = resources.ReadNetworkGroups(ctx, r, rd.ID.ValueString(), &resp.Diagnostics)
	rd.Tags = resources.ReadAddonTags(ctx, r, rd.ID.ValueString(), rd.Tags, &resp.Diagnostics)

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
	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(plan.ID.ValueString()).Updateaddoninfo(ctx, addons.UpdateaddoninfoBody{
		Name: plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Redis", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(pkg.Deref(addonRes.Payload().Name))

	addon.SyncNetworkGroups(
		ctx,
		r,
		plan.ID.ValueString(),
		plan.Networkgroups,
		&state.Networkgroups,
		&resp.Diagnostics,
	)
	if !plan.Tags.Equal(state.Tags) {
		state.Tags = resources.SyncAddonTags(ctx, r, plan.ID.ValueString(), plan.Tags, state.Tags, &resp.Diagnostics)
	}

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

	res := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(rd.ID.ValueString()).Deprovisionaddonbyorgaandaddonid(ctx)
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
