package rust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

// Create a new resource
func (r *ResourceRust) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceRust.Create()")

	plan := helper.PlanFrom[Rust](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Create(ctx, r, &plan)...)

	// First save: persist ID even if there were partial errors
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Secondary operations
	application.SyncNetworkGroups(ctx, r, plan.ID.ValueString(), plan.Networkgroups, &resp.Diagnostics)
	application.SyncExposedVariables(ctx, r, plan.ID.ValueString(), plan.ExposedEnvironment, &resp.Diagnostics)
	application.SyncDependencies(ctx, r, plan.ID.ValueString(), plan.Dependencies, &resp.Diagnostics)
	application.GitDeploy(ctx, plan.ToDeployment(r.GitAuth()), plan.DeployURL.ValueString(), &resp.Diagnostics)

	// Second save: persist secondary operations results
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceRust) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceRust.Read()")

	state := helper.StateFrom[Rust](ctx, req.State, &resp.Diagnostics)
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
func (r *ResourceRust) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Rust](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Rust](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Update(ctx, r, &plan, &state)...)

	// First save: persist changes even if there were partial errors
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Secondary operations
	application.SyncNetworkGroups(ctx, r, plan.ID.ValueString(), plan.Networkgroups, &resp.Diagnostics)
	application.SyncExposedVariables(ctx, r, plan.ID.ValueString(), plan.ExposedEnvironment, &resp.Diagnostics)
	application.SyncDependencies(ctx, r, plan.ID.ValueString(), plan.Dependencies, &resp.Diagnostics)

	// Second save: persist secondary operations results
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ResourceRust) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.PlanFrom[Rust](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	if application.DefaultAndValidateRuntimePlan(ctx, r, "rust", &plan.Runtime, &res.Diagnostics) {
		res.Diagnostics.Append(res.Plan.Set(ctx, plan)...)
	}
}
