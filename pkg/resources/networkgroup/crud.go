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
	"go.clever-cloud.dev/sdk/models"
)

// Create a new resource
func (r *ResourceNG) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[Networkgroup](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := tmp.GenID()
	plan.ID = basetypes.NewStringValue(id)

	label := plan.Name.ValueString()
	description := plan.Description.ValueString()
	ngRes := r.SDK.
		V4().
		Networkgroups().
		Organisations().
		Ownerid(r.Organization()).
		Networkgroups().
		Createnetworkgroup(ctx, &models.WannabeNetworkGroup{
			ID:          &id,
			Label:       &label,
			Description: &description,
			Tags:        pkg.SetToStringSlice(ctx, plan.Tags, &resp.Diagnostics),
		})
	if ngRes.HasError() {
		resp.Diagnostics.AddError("failed to create networkgroup", ngRes.Error().Error())
		return
	}

	ng, err := r.WaitForNG(ctx, id)
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

	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	ngRes := r.SDK.
		V4().
		Networkgroups().
		Organisations().
		Ownerid(r.Organization()).
		Networkgroups().
		Networkgroupid(state.ID.ValueString()).
		Getnetworkgroup(ctx)
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

	res := r.SDK.
		V4().
		Networkgroups().
		Organisations().
		Ownerid(r.Organization()).
		Networkgroups().
		Networkgroupid(state.ID.ValueString()).
		Deletenetworkgroup(ctx)
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete networkgroup", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ResourceNG) WaitForNG(ctx context.Context, ngId string) (*models.NetworkGroup1, error) {
	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return nil, lastErr
		default:
			res := r.SDK.
				V4().
				Networkgroups().
				Organisations().
				Ownerid(r.Organization()).
				Networkgroups().
				Networkgroupid(ngId).
				Getnetworkgroup(ctx)
			if res.HasError() {
				lastErr = res.Error()
				time.Sleep(1 * time.Second)
				continue
			}

			return res.Payload(), nil
		}
	}

}
