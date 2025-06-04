package static

import (
	"context"
	"reflect"

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
func (r *ResourceStatic) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceStatic.Configure()")

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
func (r *ResourceStatic) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := Static{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstanceByVariantSlug(ctx, r.cc, nil, "static-apache", resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vhosts := []string{}
	resp.Diagnostics.Append(plan.AdditionalVHosts.ElementsAs(ctx, &vhosts, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment := plan.toEnv(ctx, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createAppReq := application.CreateReq{
		Client:       r.cc,
		Organization: r.org,
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
		Environment: environment,
		VHosts:      vhosts,
		Deployment:  plan.toDeployment(),
	}

	createAppRes, diags := application.CreateApp(ctx, createAppReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "BUILD FLAVOR RES"+createAppRes.Application.BuildFlavor.Name, map[string]any{})
	plan.ID = pkg.FromStr(createAppRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createAppRes.Application.DeployURL)
	plan.VHost = pkg.FromStr(createAppRes.Application.Vhosts[0].Fqdn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceStatic) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Static

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readRes, diags := application.ReadApp(ctx, r.cc, r.org, state.ID.ValueString())
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

	vhosts := readRes.App.Vhosts.AsString()
	hasDefaultVHost := pkg.HasSome(vhosts, func(vhost string) bool {
		return pkg.VhostCleverAppsRegExp.MatchString(vhost)
	})
	if hasDefaultVHost {
		cleverapps := *pkg.First(vhosts, func(vhost string) bool {
			return pkg.VhostCleverAppsRegExp.MatchString(vhost)
		})
		state.VHost = pkg.FromStr(cleverapps)
	} else {
		state.VHost = types.StringNull()
	}

	vhostsWithoutDefault := pkg.Filter(vhosts, func(vhost string) bool {
		ok := pkg.VhostCleverAppsRegExp.MatchString(vhost)
		return !ok
	})
	if len(vhostsWithoutDefault) > 0 {
		state.AdditionalVHosts = pkg.FromListString(vhostsWithoutDefault)
	} else {
		state.AdditionalVHosts = types.ListNull(types.StringType)
	}

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
func (r *ResourceStatic) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceStatic.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Static](ctx, req.Plan, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Static](ctx, req.State, res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Retrieve instance of the app from context
	instance := application.LookupInstanceByVariantSlug(ctx, r.cc, nil, "static-apache", res.Diagnostics)
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
		Environment:    planEnvironment,
		VHosts:         vhosts,
		Deployment:     plan.toDeployment(),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
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
func (r *ResourceStatic) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Static

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "STATIC DELETE", map[string]any{"state": state})

	res := tmp.DeleteApp(ctx, r.cc, r.org, state.ID.ValueString())
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
func (r *ResourceStatic) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
