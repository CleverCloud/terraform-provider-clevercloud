package scala

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceScala) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := Scala{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "sbt", resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &resp.Diagnostics)

	environment := plan.toEnv(ctx, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
		Deployment:   plan.toDeployment(r.GitAuth()),
		Dependencies: dependencies,
	}

	createAppRes, diags := application.CreateApp(ctx, createAppReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "BUILD FLAVOR RES"+createAppRes.Application.BuildFlavor.Name, map[string]any{})
	plan.ID = pkg.FromStr(createAppRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createAppRes.Application.DeployURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	createdVhosts := createAppRes.Application.Vhosts
	if plan.VHosts.IsUnknown() { // practitionner does not provide any vhost, return the cleverapps one
		plan.VHosts = helper.VHostsFromAPIHosts(createdVhosts.AsString(), &resp.Diagnostics)
	} else { // practitionner give it's own vhost, remove cleverapps one

		for _, vhost := range pkg.Diff(vhosts, createdVhosts.AsString()) {
			deleteVhostRes := tmp.DeleteAppVHost(
				ctx,
				r.Client(),
				r.Organization(),
				plan.ID.ValueString(),
				vhost,
			)
			if deleteVhostRes.HasError() {
				diags.AddError("failed to remove vhost", deleteVhostRes.Error().Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceScala) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Scala

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readRes, diags := application.ReadApp(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readRes.AppIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = pkg.FromStr(readRes.App.Name)
	state.Description = pkg.FromStr(readRes.App.Description)
	state.MinInstanceCount = pkg.FromI(int64(readRes.App.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(readRes.App.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(readRes.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(readRes.App.Instance.MaxFlavor.Name)
	state.Region = pkg.FromStr(readRes.App.Zone)
	state.DeployURL = pkg.FromStr(readRes.App.DeployURL)
	state.BuildFlavor = readRes.GetBuildFlavor()

	state.VHosts = helper.VHostsFromAPIHosts(readRes.App.Vhosts.AsString(), &resp.Diagnostics)

	for envName, envValue := range readRes.EnvAsMap() {
		switch envName {
		case "APP_FOLDER":
			state.AppFolder = pkg.FromStr(envValue)
		default:
			//state.Environment.
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceScala) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceScala.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Scala](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Scala](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "sbt", res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retriev all env values by extracting ctx env viriables and merge it with the app env variables
	planEnvironment := plan.toEnv(ctx, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	stateEnvironment := state.toEnv(ctx, res.Diagnostics)
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
	updatedApp, diags := application.UpdateApp(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = helper.VHostsFromAPIHosts(updatedApp.Application.Vhosts.AsString(), &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourceScala) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Scala

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "SCALA DELETE", map[string]any{"state": state})

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

// Import resource
