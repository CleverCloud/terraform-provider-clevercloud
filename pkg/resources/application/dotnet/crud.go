package dotnet

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceDotnet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Dotnet](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &resp.Diagnostics)
	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "dotnet", &resp.Diagnostics)
	environment := plan.toEnv(ctx, &resp.Diagnostics)
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
		Dependencies: dependencies,
	}

	createRes, diags := application.CreateApp(ctx, createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createRes.Application.DeployURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	createdVhosts := createRes.Application.Vhosts
	plan.VHosts = helper.VHostsFromAPIHosts(ctx, createdVhosts.AsString(), plan.VHosts, &resp.Diagnostics)
	plan.StickySessions = pkg.FromBool(createRes.Application.StickySessions)
	plan.RedirectHTTPS = pkg.FromBool(application.ToForceHTTPS(createRes.Application.ForceHTTPS))

	application.SyncNetworkGroups(
		ctx,
		r.Client(),
		r.Organization(),
		createRes.Application.ID,
		plan.Networkgroups,
		&resp.Diagnostics,
	)

	deploy := plan.toDeployment(r.GitAuth())
	if deploy != nil {
		application.GitDeploy(ctx, *deploy, createRes.Application.DeployURL, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceDotnet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Dotnet READ", map[string]any{"request": req})

	state := helper.StateFrom[Dotnet](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	appRes, diags := application.ReadApp(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if appRes.AppIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.DeployURL = pkg.FromStr(appRes.App.DeployURL)
	state.Name = pkg.FromStr(appRes.App.Name)
	state.Description = pkg.FromStr(appRes.App.Description)
	state.Region = pkg.FromStr(appRes.App.Zone)
	state.MinInstanceCount = basetypes.NewInt64Value(int64(appRes.App.Instance.MinInstances))
	state.MaxInstanceCount = basetypes.NewInt64Value(int64(appRes.App.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(appRes.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appRes.App.Instance.MaxFlavor.Name)
	state.BuildFlavor = appRes.GetBuildFlavor()
	//state.DotnetProfile = pkg.FromStr(appRes.App.DotnetProfile)
	//state.DotnetProj = pkg.FromStr(appRes.App.DotnetProj)
	//state.DotnetTFM = pkg.FromStr(appRes.App.DotnetTFM)
	//state.DotnetVersion = pkg.FromStr(appRes.App.DotnetVersion)

	state.VHosts = helper.VHostsFromAPIHosts(ctx, appRes.App.Vhosts.AsString(), state.VHosts, &resp.Diagnostics)
	state.Networkgroups = resources.ReadNetworkGroups(ctx, r.Client(), r.Organization(), state.ID.ValueString(), &resp.Diagnostics)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceDotnet) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceDotnet.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Dotnet](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Dotnet](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "dotnet", &res.Diagnostics)
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

	updatedApp, diags := application.UpdateApp(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = helper.VHostsFromAPIHosts(ctx, updatedApp.Application.Vhosts.AsString(), plan.VHosts, &res.Diagnostics)

	application.SyncNetworkGroups(
		ctx,
		r.Client(),
		r.Organization(),
		state.ID.ValueString(),
		plan.Networkgroups,
		&res.Diagnostics,
	)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourceDotnet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Dotnet](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Dotnet DELETE", map[string]any{"app": state.ID.ValueString()})

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
