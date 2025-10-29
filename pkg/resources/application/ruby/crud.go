package ruby

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/helper/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceRuby) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Ruby](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &resp.Diagnostics)

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "ruby", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	environment := plan.toEnv(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dependencies := plan.DependenciesAsString(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := application.CreateReq{
		Client:       r.Client(),
		Organization: r.Organization(),
		Application: tmp.CreateAppRequest{
			Name:            plan.Name.ValueString(),
			Deploy:          "git",
			Description:     plan.Description.ValueString(),
			InstanceType:    instance.Type,
			InstanceVariant: instance.Variant.ID,
			InstanceVersion: instance.Version,
			MinFlavor:       plan.SmallestFlavor.ValueString(),
			MaxFlavor:       plan.BiggestFlavor.ValueString(),
			MinInstances:    plan.MinInstanceCount.ValueInt64(),
			MaxInstances:    plan.MaxInstanceCount.ValueInt64(),
			BuildFlavor:     plan.BuildFlavor.ValueString(),
			StickySessions:  plan.StickySessions.ValueBool(),
			ForceHttps:      application.FromForceHTTPS(plan.RedirectHTTPS.ValueBool()),
			Zone:            plan.Region.ValueString(),
		},
		Environment:  environment,
		VHosts:       vhosts,
		Deployment:   plan.toDeployment(r.GitAuth()),
		Dependencies: dependencies,
	}

	createRes, diags := application.Create(ctx, createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createRes.Application.DeployURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	createdVhosts := createRes.Application.Vhosts
	plan.VHosts = helper.VHostsFromAPIHosts(ctx, createdVhosts.AsString(), plan.VHosts, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceRuby) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Ruby READ", map[string]any{"request": req})

	state := helper.StateFrom[Ruby](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	appRes, diags := application.Read(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if appRes.AppIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.DeployURL = pkg.FromStr(appRes.App.DeployURL)

	state.VHosts = helper.VHostsFromAPIHosts(ctx, appRes.App.Vhosts.AsString(), state.VHosts, &resp.Diagnostics)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceRuby) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceRuby.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Ruby](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Ruby](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "ruby", &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retriev all env values by extracting ctx env viriables and merge it with the app env variables
	planEnvironment := plan.toEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	stateEnvironment := state.toEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Same as env but with vhosts
	vhosts := plan.VHostsAsStrings(ctx, &res.Diagnostics)

	// Get the updated values from plan and instance
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
		Deployment:     plan.toDeployment(r.GitAuth()),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
	}

	// Correctly named: update the app (via PUT Method)
	updatedApp, diags := application.Update(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = helper.VHostsFromAPIHosts(ctx, updatedApp.Application.Vhosts.AsString(), plan.VHosts, &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourceRuby) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Ruby](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Ruby DELETE", map[string]any{"app": state.ID.ValueString()})

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
