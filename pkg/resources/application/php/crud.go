package php

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourcePHP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourcePHP.Create()")
	plan := helper.PlanFrom[PHP](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "php", &resp.Diagnostics)
	vhosts := plan.VHostsAsStrings(ctx, &resp.Diagnostics)
	environment := plan.toEnv(ctx, &resp.Diagnostics)
	dependencies := plan.DependenciesAsString(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createAppReq := application.CreateReq{
		Client:       r.Client(),
		Organization: r.Organization(),
		Application: tmp.CreateAppRequest{
			Name:            plan.Name.ValueString(),
			Deploy:          "git",
			Description:     plan.Description.ValueString(),
			InstanceType:    instance.Type,
			InstanceVariant: instance.Variant.ID,
			InstanceVersion: instance.Version,
			BuildFlavor:     plan.BuildFlavor.ValueString(),
			MinFlavor:       plan.SmallestFlavor.ValueString(),
			MaxFlavor:       plan.BiggestFlavor.ValueString(),
			MinInstances:    plan.MinInstanceCount.ValueInt64(),
			MaxInstances:    plan.MaxInstanceCount.ValueInt64(),
			StickySessions:  plan.StickySessions.ValueBool(),
			ForceHttps:      application.FromForceHTTPS(plan.RedirectHTTPS.ValueBool()),
			Zone:            plan.Region.ValueString(),
			CancelOnPush:    false,
		},
		Environment:  environment,
		VHosts:       vhosts,
		Dependencies: dependencies,
	}

	createAppRes, diags := application.CreateApp(ctx, createAppReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createAppRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createAppRes.Application.DeployURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	createdVhosts := createAppRes.Application.Vhosts
	plan.VHosts = helper.VHostsFromAPIHosts(ctx, createdVhosts.AsString(), plan.VHosts, &resp.Diagnostics)
	plan.StickySessions = pkg.FromBool(createAppRes.Application.StickySessions)
	plan.RedirectHTTPS = pkg.FromBool(application.ToForceHTTPS(createAppRes.Application.ForceHTTPS))

	deploy := plan.toDeployment(r.GitAuth())
	if deploy != nil {
		application.GitDeploy(ctx, *deploy, createAppRes.Application.DeployURL, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourcePHP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourcePHP.Read()")
	state := helper.StateFrom[PHP](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	appPHP, diags := application.ReadApp(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if appPHP.AppIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = pkg.FromStr(appPHP.App.Name)
	state.Description = pkg.FromStr(appPHP.App.Description)
	state.MinInstanceCount = pkg.FromI(int64(appPHP.App.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(appPHP.App.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(appPHP.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appPHP.App.Instance.MaxFlavor.Name)
	state.Region = pkg.FromStr(appPHP.App.Zone)
	state.DeployURL = pkg.FromStr(appPHP.App.DeployURL)
	state.BuildFlavor = appPHP.GetBuildFlavor()

	state.VHosts = helper.VHostsFromAPIHosts(ctx, appPHP.App.Vhosts.AsString(), state.VHosts, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourcePHP) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourcePHP.Update()")

	plan := helper.PlanFrom[PHP](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[PHP](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "php", &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	planEnvironment := plan.toEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	stateEnvironment := state.toEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &res.Diagnostics)
	dependencies := plan.DependenciesAsString(ctx, &res.Diagnostics)

	updateAppReq := application.UpdateReq{
		ID:           state.ID.ValueString(),
		Client:       r.Client(),
		Organization: r.Organization(),
		Application: tmp.UpdateAppReq{
			Name:            plan.Name.ValueString(),
			Deploy:          "git",
			Description:     plan.Description.ValueString(),
			InstanceType:    instance.Type,
			InstanceVariant: instance.Variant.ID,
			InstanceVersion: instance.Version,
			BuildFlavor:     plan.BuildFlavor.ValueString(),
			MinFlavor:       plan.SmallestFlavor.ValueString(),
			MaxFlavor:       plan.BiggestFlavor.ValueString(),
			MinInstances:    plan.MinInstanceCount.ValueInt64(),
			MaxInstances:    plan.MaxInstanceCount.ValueInt64(),
			StickySessions:  plan.StickySessions.ValueBool(),
			ForceHttps:      application.FromForceHTTPS(plan.RedirectHTTPS.ValueBool()),
			Zone:            plan.Region.ValueString(),
			CancelOnPush:    false,
		},
		Environment:    planEnvironment,
		VHosts:         vhosts,
		Dependencies:   dependencies,
		Deployment:     plan.toDeployment(r.GitAuth()),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
	}

	updatedApp, diags := application.UpdateApp(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = helper.VHostsFromAPIHosts(ctx, updatedApp.Application.Vhosts.AsString(), plan.VHosts, &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourcePHP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourcePHP.Delete()")
	var state PHP

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "PHP DELETE", map[string]any{"state": state})

	res := tmp.DeleteApp(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete app", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
