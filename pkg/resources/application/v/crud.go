package v

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
func (r *ResourceV) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceV.Create()")

	plan := helper.PlanFrom[V](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.GenericCreate(ctx, r, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceV) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "V READ", map[string]any{"request": req})

	state := helper.StateFrom[V](ctx, req.State, &resp.Diagnostics)
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
	state.Name = pkg.FromStr(appRes.App.Name)
	state.Description = pkg.FromStr(appRes.App.Description)
	state.Region = pkg.FromStr(appRes.App.Zone)
	state.MinInstanceCount = basetypes.NewInt64Value(int64(appRes.App.Instance.MinInstances))
	state.MaxInstanceCount = basetypes.NewInt64Value(int64(appRes.App.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(appRes.App.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appRes.App.Instance.MaxFlavor.Name)

	state.VHosts = helper.VHostsFromAPIHosts(ctx, appRes.App.Vhosts.AsString(), state.VHosts, &resp.Diagnostics)
	state.Networkgroups = resources.ReadNetworkGroups(ctx, r, state.ID.ValueString(), &resp.Diagnostics)
	state.ExposedEnvironment = application.ReadExposedVariables(ctx, r, appRes.App.ID, &resp.Diagnostics)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceV) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceV.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[V](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[V](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstanceByVariantSlug(ctx, r.Client(), nil, "v", &res.Diagnostics)
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
		Deployment:     plan.ToDeployment(r.GitAuth()),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
	}

	updatedApp, diags := application.Update(ctx, updateAppReq)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}

	plan.VHosts = helper.VHostsFromAPIHosts(ctx, updatedApp.Application.Vhosts.AsString(), plan.VHosts, &res.Diagnostics)

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
func (r *ResourceV) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[V](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "V DELETE", map[string]any{"app": state.ID.ValueString()})

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

func (r *ResourceV) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.PlanFrom[V](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	application.ValidateRuntimeFlavors(ctx, r, "v", plan.Runtime, &res.Diagnostics)
}
