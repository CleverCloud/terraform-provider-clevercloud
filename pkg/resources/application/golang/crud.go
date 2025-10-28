package golang

import (
	"context"
	"reflect"

	"go.clever-cloud.com/terraform-provider/pkg/helper/application"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceGo) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	tflog.Debug(ctx, "Go CREATE")

	plan := helper.PlanFrom[Go](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &res.Diagnostics)

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "go", &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	environment := plan.toEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	dependencies := plan.DependenciesAsString(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
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
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createRes.Application.DeployURL)
	plan.BuildFlavor = createRes.GetBuildFlavor()

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)

	createdVhosts := createRes.Application.Vhosts
	plan.VHosts = helper.VHostsFromAPIHosts(ctx, createdVhosts.AsString(), plan.VHosts, &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceGo) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	tflog.Debug(ctx, "Go READ", map[string]any{"request": req})

	state := helper.StateFrom[Go](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	appRes, diags := application.Read(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	if appRes.AppIsDeleted {
		res.State.RemoveResource(ctx)
		return
	}

	state.DeployURL = pkg.FromStr(appRes.App.DeployURL)
	state.Name = pkg.FromStr(appRes.App.Name)
	state.Description = pkg.FromStr(appRes.App.Description)
	state.Region = pkg.FromStr(appRes.App.Zone)
	state.SmallestFlavor = pkg.FromStr(appRes.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appRes.App.Instance.MaxFlavor.Name)
	state.MinInstanceCount = pkg.FromI(int64(appRes.App.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(appRes.App.Instance.MaxInstances))
	state.BuildFlavor = appRes.GetBuildFlavor()
	state.StickySessions = pkg.FromBool(appRes.App.StickySessions)
	state.RedirectHTTPS = pkg.FromBool(application.ToForceHTTPS(appRes.App.ForceHTTPS))
	state.VHosts = helper.VHostsFromAPIHosts(ctx, appRes.App.Vhosts.AsString(), state.VHosts, &res.Diagnostics)

	diags = res.State.Set(ctx, state)
	res.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceGo) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceGo.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Go](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Go](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "go", &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve all env values by extracting ctx env variables and merge it with the app env variables
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
	dependencies := plan.DependenciesAsString(ctx, &res.Diagnostics)

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
		Dependencies:   dependencies,
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
func (r *ResourceGo) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	state := helper.StateFrom[Go](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Go DELETE", map[string]any{"state": state})

	deleteRes := tmp.DeleteApp(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if deleteRes.IsNotFoundError() {
		res.State.RemoveResource(ctx)
		return
	}
	if deleteRes.HasError() {
		res.Diagnostics.AddError("failed to delete app", deleteRes.Error().Error())
		return
	}

	res.State.RemoveResource(ctx)
}
