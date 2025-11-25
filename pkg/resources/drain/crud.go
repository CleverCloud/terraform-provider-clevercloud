package drain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.dev/sdk/models"
)

// Create drain resource
func (r *ResourceDrain[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[T](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the drain via SDK
	recipient := plan.ToSDKRecipient()
	wannabeDrain := models.WannabeDrain{
		Kind:      models.DrainKind(plan.GetDrain().Kind.ValueString()),
		Recipient: recipient,
	}

	createRes := r.
		SDK().
		V4().
		Drains().
		Organisations().
		Ownerid(r.Organization()).
		Applications().
		Applicationid(plan.GetDrain().ResourceID.ValueString()).
		Drains().
		Createdrain(ctx, &wannabeDrain)
	if createRes.HasError() {
		resp.Diagnostics.AddError("Failed to create drain", createRes.Error().Error())
		return
	}
	drain := createRes.Payload()

	// Update state with API response data
	state := plan
	// First update the common drain fields
	state.SetDrain(Drain{
		ID:         types.StringValue(drain.ID),
		Kind:       types.StringValue(string(drain.Kind)),
		ResourceID: types.StringValue(drain.ResourceID),
	})
	// Then update the recipient-specific fields from the API response
	state.FromSDKRecipient(drain.Recipient)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read drain resource
func (r *ResourceDrain[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[T](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get drain from API via SDK
	drainRes := r.
		SDK().
		V4().
		Drains().
		Organisations().
		Ownerid(r.Organization()).
		Applications().
		Applicationid(state.GetDrain().ResourceID.ValueString()).
		Drains().
		Drainid(state.GetDrain().ID.ValueString()).
		Getdrain(ctx)
	if drainRes.HasError() {
		resp.Diagnostics.AddError("Failed to read drain", drainRes.Error().Error())
		return
	}
	drain := drainRes.Payload()

	// Update common drain fields
	state.SetDrain(Drain{
		ID:         types.StringValue(drain.ID),
		Kind:       types.StringValue(string(drain.Kind)),
		ResourceID: types.StringValue(drain.ResourceID),
	})
	// Update recipient-specific fields from API (non-sensitive fields only)
	// Sensitive fields are automatically preserved by Terraform when marked as Sensitive: true
	state.FromSDKRecipient(drain.Recipient)

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

	deleteRes := r.
		SDK().
		V4().
		Drains().
		Organisations().
		Ownerid(r.Organization()).
		Applications().
		Applicationid(state.GetDrain().ResourceID.ValueString()).
		Drains().
		Drainid(state.GetDrain().ID.ValueString()).
		Deletedrain(ctx)
	if !deleteRes.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.AddError("Failed to delete drain", deleteRes.Error().Error())
}
