package rust

import (
	"context"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceRust) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceRust.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
		r.gitAuth = provider.GitAuth()
	}

	tflog.Debug(ctx, "AFTER CONFIGURED", map[string]any{"cc": r.cc == nil, "org": r.org})
}

// Create a new resource
func (r *ResourceRust) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	plan := helper.PlanFrom[Rust](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &res.Diagnostics)

	instance := application.LookupInstanceByVariantSlug(ctx, r.cc, nil, "rust", res.Diagnostics)
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
		Client:       r.cc,
		Organization: r.org,
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
		Deployment:   plan.toDeployment(r.gitAuth),
		Dependencies: dependencies,
	}

	createRes, diags := application.CreateApp(ctx, createReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createRes.Application.DeployURL)
	plan.BuildFlavor = createRes.GetBuildFlavor()

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)

	createdVhosts := createRes.Application.Vhosts
	if plan.VHosts.IsUnknown() { // practitionner does not provide any vhost, return the cleverapps one
		plan.VHosts = pkg.FromSetString(createdVhosts.AsString(), &res.Diagnostics)
	} else { // practitionner give it's own vhost, remove cleverapps one

		for _, vhost := range pkg.Diff(vhosts, createdVhosts.AsString()) {
			deleteVhostRes := tmp.DeleteAppVHost(
				ctx,
				r.cc,
				r.org,
				plan.ID.ValueString(),
				vhost,
			)
			if deleteVhostRes.HasError() {
				diags.AddError("failed to remove vhost", deleteVhostRes.Error().Error())
				return
			}
		}
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceRust) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	state := helper.StateFrom[Rust](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	appRes, diags := application.ReadApp(ctx, r.cc, r.org, state.ID.ValueString())
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

	vhosts := appRes.App.Vhosts.AsString()
	state.VHosts = pkg.FromSetString(vhosts, &res.Diagnostics)

	if env := appRes.EnvAsMap(); env[CC_RUST_FEATURES] != "" {
		state.Features = pkg.FromSetString(strings.Split(env[CC_RUST_FEATURES], ","), &res.Diagnostics)
	}

	diags = res.State.Set(ctx, state)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceRust) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	plan := helper.PlanFrom[Rust](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Rust](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	vhosts := plan.VHostsAsStrings(ctx, &res.Diagnostics)
	dependencies := plan.DependenciesAsString(ctx, &res.Diagnostics)

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.cc, nil, "rust", res.Diagnostics)
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

	updateAppReq := application.UpdateReq{
		ID:           state.ID.ValueString(),
		Client:       r.cc,
		Organization: r.org,
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
		Deployment:     plan.toDeployment(r.gitAuth),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
	}

	// Correctly named: update the app (via PUT Method)
	updatedApp, diags := application.UpdateApp(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = pkg.FromSetString(updatedApp.Application.Vhosts.AsString(), &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourceRust) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	state := helper.StateFrom[Rust](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	deleteRes := tmp.DeleteApp(ctx, r.cc, r.org, state.ID.ValueString())
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

// Import resource
func (r *ResourceRust) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
