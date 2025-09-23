package drain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create drain resource
func (r *ResourceDrain[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[T](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the drain via API
	wannabeDrain := tmp.WannabeDrain{
		Kind:      tmp.DRAIN_KIND(plan.GetDrain().Kind.ValueString()),
		Recipient: plan.ToRecipient(),
	}

	createRes := tmp.CreateDrain(ctx, r.Client(), r.Organization(), plan.GetDrain().ResourceID.ValueString(), wannabeDrain)
	if createRes.HasError() {
		resp.Diagnostics.AddError("Failed to create drain", createRes.Error().Error())
		return
	}
	drain := createRes.Payload()

	// Update state with API response data
	state := plan
	state.SetDrain(Drain{
		ID:         types.StringValue(drain.ID),
		Kind:       types.StringValue(string(drain.Kind)),
		ResourceID: types.StringValue(drain.ApplicationID),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read drain resource
func (r *ResourceDrain[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[T](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get drain from API
	drainRes := tmp.GetDrain(ctx, r.Client(), r.Organization(), state.GetDrain().ResourceID.ValueString(), state.GetDrain().ID.ValueString())
	if drainRes.HasError() {
		resp.Diagnostics.AddError("Failed to read drain", drainRes.Error().Error())
		return
	}
	drain := drainRes.Payload()

	// Update state from API data while preserving sensitive values
	err := state.FromAPI(*drain)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse drain data", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update drain resource
func (r *ResourceDrain[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Drains cannot be updated, need to be recreated
	resp.Diagnostics.AddError("Update not supported", "Drains cannot be updated in place. Please recreate the resource.")
}

// Delete drain resource
func (r *ResourceDrain[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[T](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteRes := tmp.DeleteDrain(ctx, r.Client(), r.Organization(), state.GetDrain().ResourceID.ValueString(), state.GetDrain().ID.ValueString())
	if !deleteRes.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.AddError("Failed to delete drain", deleteRes.Error().Error())

}
