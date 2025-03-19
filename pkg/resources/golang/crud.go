package golang

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceGo) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceGo.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}

	tflog.Debug(ctx, "AFTER CONFIGURED", map[string]any{"cc": r.cc == nil, "org": r.org})
}

// Create a new resource
func (r *ResourceGo) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	plan := helper.PlanFrom[Go](ctx, req.Plan, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	vhosts := []string{}
	res.Diagnostics.Append(plan.AdditionalVHosts.ElementsAs(ctx, &vhosts, false)...)
	if res.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstance(ctx, r.cc, "go", "Go", res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	environment := plan.toEnv(ctx, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	dependencies := []string{}
	res.Diagnostics.Append(plan.Dependencies.ElementsAs(ctx, &dependencies, false)...)
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
			StickySessions:  plan.StickySessions.ValueBool(),
			ForceHttps:      application.FromForceHTTPS(plan.RedirectHTTPS.ValueBool()),
			Zone:            plan.Region.ValueString(),
		},
		Environment:  environment,
		VHosts:       vhosts,
		Deployment:   plan.toDeployment(),
		Dependencies: dependencies,
	}

	createRes, diags := application.CreateApp(ctx, createReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.ID = pkg.FromStr(createRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createRes.Application.DeployURL)
	plan.VHost = pkg.FromStr(createRes.Application.Vhosts[0].Fqdn)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceGo) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	tflog.Debug(ctx, "Go READ", map[string]any{"request": req})

	state := helper.StateFrom[Go](ctx, req.State, res.Diagnostics)
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
	state.VHost = pkg.FromStr(appRes.App.Vhosts[0].Fqdn)

	diags = res.State.Set(ctx, state)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceGo) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceGo.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Go](ctx, req.Plan, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Go](ctx, req.State, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstance(ctx, r.cc, "go", "Go", res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retriev all env values by extracting ctx env viriables and merge it with the app env variables
	environment := plan.toEnv(ctx, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Same as env but with vhosts
	vhosts := []string{}
	if res.Diagnostics.Append(plan.AdditionalVHosts.ElementsAs(ctx, &vhosts, false)...); res.Diagnostics.HasError() {
		return
	}

	// Get the updated values from plan and instance
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
		Environment: environment,
		VHosts:      vhosts,
		Deployment:  plan.toDeployment(),
	}

	// Correctly named: update the app (via PUT Method)
	_, diags := application.UpdateApp(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	//
	hasDefaultVHost := pkg.HasSome(updateAppReq.VHosts, func(vhost string) bool {
		return pkg.VhostCleverAppsRegExp.MatchString(vhost)
	})
	if hasDefaultVHost {
		cleverapps := *pkg.First(vhosts, func(vhost string) bool {
			return pkg.VhostCleverAppsRegExp.MatchString(vhost)
		})
		plan.VHost = pkg.FromStr(cleverapps)
	} else {
		plan.VHost = types.StringNull()
	}

	vhostsWithoutDefault := pkg.Filter(updateAppReq.VHosts, func(vhost string) bool {
		ok := pkg.VhostCleverAppsRegExp.MatchString(vhost)
		return !ok
	})
	if len(vhostsWithoutDefault) > 0 {
		plan.AdditionalVHosts = pkg.FromListString(vhostsWithoutDefault)
	} else {
		plan.AdditionalVHosts = types.ListNull(types.StringType)
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourceGo) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	state := helper.StateFrom[Go](ctx, req.State, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Go DELETE", map[string]any{"state": state})

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
func (r *ResourceGo) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
