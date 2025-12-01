package frankenphp

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceFrankenPHP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceFrankenPHP.Create()")

	plan := helper.PlanFrom[FrankenPHP](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Create(ctx, r, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceFrankenPHP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceFrankenPHP.Read()")
	state := helper.StateFrom[FrankenPHP](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	appFrankenPHP, diags := application.Read(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if appFrankenPHP.AppIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = pkg.FromStr(appFrankenPHP.App.Name)
	state.Description = pkg.FromStr(appFrankenPHP.App.Description)
	state.MinInstanceCount = pkg.FromI(int64(appFrankenPHP.App.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(appFrankenPHP.App.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(appFrankenPHP.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appFrankenPHP.App.Instance.MaxFlavor.Name)
	state.Region = pkg.FromStr(appFrankenPHP.App.Zone)
	state.DeployURL = pkg.FromStr(appFrankenPHP.App.DeployURL)

	state.VHosts = helper.VHostsFromAPIHosts(ctx, appFrankenPHP.App.Vhosts.AsString(), state.VHosts, &resp.Diagnostics)
	state.Networkgroups = resources.ReadNetworkGroups(ctx, r, state.ID.ValueString(), &resp.Diagnostics)
	state.ExposedEnvironment = application.ReadExposedVariables(ctx, r, appFrankenPHP.App.ID, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceFrankenPHP) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceFrankenPHP.Update()")

	plan := helper.PlanFrom[FrankenPHP](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[FrankenPHP](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "frankenphp", &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	planEnvironment := plan.ToEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	stateEnvironment := state.ToEnv(ctx, &res.Diagnostics)
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
		Deployment:     plan.ToDeployment(r.GitAuth()),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
	}

	updateAppRes, diags := application.Update(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = helper.VHostsFromAPIHosts(ctx, updateAppRes.Application.Vhosts.AsString(), plan.VHosts, &res.Diagnostics)

	plan.ID = pkg.FromStr(updateAppRes.Application.ID)
	plan.DeployURL = pkg.FromStr(updateAppRes.Application.DeployURL)

	application.SyncNetworkGroups(
		ctx,
		r,
		state.ID.ValueString(),
		plan.Networkgroups,
		&res.Diagnostics,
	)

	application.SyncExposedVariables(
		ctx,
		r,
		state.ID.ValueString(),
		plan.ExposedEnvironment,
		&res.Diagnostics,
	)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourceFrankenPHP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourceFrankenPHP.Delete()")
	state := helper.StateFrom[FrankenPHP](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteAppRes := tmp.DeleteApp(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if deleteAppRes.HasError() {
		resp.Diagnostics.AddError("failed to delete app", deleteAppRes.Error().Error())
	}
}

func (r *ResourceFrankenPHP) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.PlanFrom[FrankenPHP](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	application.ValidateRuntimeFlavors(ctx, r, "frankenphp", plan.Runtime, &res.Diagnostics)
}
