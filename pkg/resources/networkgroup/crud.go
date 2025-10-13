package networkgroup

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// Create a new resource
func (r *ResourceNG) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Networkgroup](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := tmp.GenID()
	plan.ID = basetypes.NewStringValue(id)

	ngRes := tmp.CreateNetworkgroup(ctx, r.Client(), r.Organization(), tmp.NetworkgroupCreation{
		ID:          id,
		Label:       plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tags:        pkg.SetToStringSlice(ctx, plan.Tags, &resp.Diagnostics),
	})
	if ngRes.HasError() {
		resp.Diagnostics.AddError("failed to create networkgroup", ngRes.Error().Error())
		return
	}

	ng, err := r.WaitForNG(ctx, r.Client(), r.Organization(), id)
	if err != nil {
		resp.Diagnostics.AddError("failed to get created networkgroup", err.Error())
		return
	}

	plan.Name = basetypes.NewStringValue(ng.Label)
	plan.Description = basetypes.NewStringPointerValue(ng.Description)
	plan.Tags = pkg.FromSetString(ng.Tags, &resp.Diagnostics)
	plan.Network = pkg.FromStr(ng.NetworkIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourceNG) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceNG READ", map[string]any{"request": req})

	state := helper.StateFrom[Networkgroup](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ngRes := tmp.GetNetworkgroup(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if ngRes.HasError() {
		resp.Diagnostics.AddError("failed to get networkgroup", ngRes.Error().Error())
		return
	}
	ng := ngRes.Payload()

	state.Name = basetypes.NewStringValue(ng.Label)
	state.Description = basetypes.NewStringPointerValue(ng.Description)
	state.Tags = pkg.FromSetString(ng.Tags, &resp.Diagnostics)
	state.Network = pkg.FromStr(ng.NetworkIP)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceNG) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("update not supported", "update not supported")
}

// Delete resource
func (r *ResourceNG) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Networkgroup DELETE")

	state := helper.StateFrom[Networkgroup](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := tmp.DeleteNetworkgroup(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete networkgroup", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ResourceNG) WaitForNG(ctx context.Context, cc *client.Client, organisationID, ngId string) (*tmp.Networkgroup, error) {
	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return nil, lastErr
		default:
			res := tmp.GetNetworkgroup(ctx, cc, organisationID, ngId)
			if res.HasError() {
				lastErr = res.Error()
				time.Sleep(1 * time.Second)
				continue
			}

			return res.Payload(), nil
		}
	}

}
