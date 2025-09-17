package python

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
func (r *ResourcePython) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Python](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &resp.Diagnostics)

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "python", resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	environment := plan.toEnv(ctx, resp.Diagnostics)
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

	createRes, diags := application.CreateApp(ctx, createReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createRes.Application.DeployURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	createdVhosts := createRes.Application.Vhosts
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
func (r *ResourcePython) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Python READ", map[string]any{"request": req})

	state := helper.StateFrom[Python](ctx, req.State, &resp.Diagnostics)
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
	state.SmallestFlavor = pkg.FromStr(appRes.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appRes.App.Instance.MaxFlavor.Name)
	state.MinInstanceCount = pkg.FromI(int64(appRes.App.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(appRes.App.Instance.MaxInstances))
	state.BuildFlavor = appRes.GetBuildFlavor()
	state.StickySessions = pkg.FromBool(appRes.App.StickySessions)
	state.RedirectHTTPS = pkg.FromBool(application.ToForceHTTPS(appRes.App.ForceHTTPS))

	state.VHosts = helper.VHostsFromAPIHosts(appRes.App.Vhosts.AsString(), &resp.Diagnostics)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourcePython) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourcePython.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Python](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Python](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "python", res.Diagnostics)
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

	cleverAppsVhost := updatedApp.Application.Vhosts.CleverAppsFQDN(plan.ID.ValueString())
	if plan.VHosts.IsUnknown() { // practitionner does not provide any vhost, return the cleverapps one
		plan.VHosts = helper.VHostsFromAPIHosts(updatedApp.Application.Vhosts.AsString(), &res.Diagnostics)
	} else { // practitionner give it's own vhost, remove cleverapps one
		if cleverAppsVhost != nil {
			deleteVhostRes := tmp.DeleteAppVHost(
				ctx,
				r.Client(),
				r.Organization(),
				plan.ID.ValueString(),
				cleverAppsVhost.Fqdn,
			)
			if deleteVhostRes.HasError() {
				diags.AddError("failed to remove vhost", deleteVhostRes.Error().Error())
				return
			}

			plan.VHosts = helper.VHostsFromAPIHosts(updatedApp.Application.Vhosts.WithoutCleverApps(plan.ID.ValueString()).AsString(), &res.Diagnostics)
		}
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourcePython) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var app Python

	diags := req.State.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Python DELETE", map[string]any{"app": app})

	res := tmp.DeleteApp(ctx, r.Client(), r.Organization(), app.ID.ValueString())
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
