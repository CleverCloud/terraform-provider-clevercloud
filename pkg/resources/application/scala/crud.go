package scala

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

// Create a new resource
func (r *ResourceScala) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceScala.Create()")

	plan := helper.PlanFrom[Scala](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Create(ctx, r, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Secondary operations
	application.SyncNetworkGroups(ctx, r, plan.ID.ValueString(), plan.Networkgroups, &resp.Diagnostics)
	application.SyncExposedVariables(ctx, r, plan.ID.ValueString(), plan.ExposedEnvironment, &resp.Diagnostics)
	application.GitDeploy(ctx, plan.ToDeployment(r.GitAuth()), plan.DeployURL.ValueString(), &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceScala) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceScala.Read()")

	state := helper.StateFrom[Scala](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	appIsDeleted, diags := application.Read(ctx, r, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if appIsDeleted {
		resp.State.RemoveResource(ctx)
		return
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

	res.Diagnostics.Append(application.Update(ctx, r, &plan, &state)...)
	if res.Diagnostics.HasError() {
		return
	}

	// Secondary operations
	application.SyncNetworkGroups(ctx, r, plan.ID.ValueString(), plan.Networkgroups, &res.Diagnostics)
	application.SyncExposedVariables(ctx, r, plan.ID.ValueString(), plan.ExposedEnvironment, &res.Diagnostics)

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourceScala) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourceScala.Delete()")

	state := helper.StateFrom[Scala](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Delete(ctx, r, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ResourceScala) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.PlanFrom[Scala](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	application.ValidateRuntimeFlavors(ctx, r, "sbt", plan.Runtime, &res.Diagnostics)
}
