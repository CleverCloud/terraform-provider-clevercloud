package php

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

// Create a new resource
func (r *ResourcePHP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourcePHP.Create()")

	plan := helper.PlanFrom[PHP](ctx, req.Plan, &resp.Diagnostics)
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
func (r *ResourcePHP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourcePHP.Read()")

	state := helper.StateFrom[PHP](ctx, req.State, &resp.Diagnostics)
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
func (r *ResourcePHP) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	plan := helper.PlanFrom[PHP](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[PHP](ctx, req.State, &res.Diagnostics)
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
func (r *ResourcePHP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourcePHP.Delete()")

	state := helper.StateFrom[PHP](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(application.Delete(ctx, r, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ResourcePHP) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.PlanFrom[PHP](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	application.ValidateRuntimeFlavors(ctx, r, "php", plan.Runtime, &res.Diagnostics)
}
