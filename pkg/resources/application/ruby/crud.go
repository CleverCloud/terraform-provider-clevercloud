package ruby

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

// Create a new resource
func (r *ResourceRuby) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceRuby.Create()")

	plan := helper.PlanFrom[Ruby](ctx, req.Plan, &resp.Diagnostics)
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
func (r *ResourceRuby) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceRuby.Read()")

	state := helper.StateFrom[Ruby](ctx, req.State, &resp.Diagnostics)
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
func (r *ResourceRuby) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceRuby.Update()")

	// Retrieve values from plan and state
	plan := helper.PlanFrom[Ruby](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[Ruby](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	res.Diagnostics.Append(application.Update(ctx, r, &plan, &state)...)
	if res.Diagnostics.HasError() {
		return
	}

	res.Diagnostics.Append(res.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourceRuby) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourceRuby.Delete()")

	state := helper.StateFrom[Ruby](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Delete(ctx, r, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ResourceRuby) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.PlanFrom[Ruby](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	application.ValidateRuntimeFlavors(ctx, r, "ruby", plan.Runtime, &res.Diagnostics)
}
