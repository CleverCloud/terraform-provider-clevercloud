package networkgroup

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

// defaultTagsFor returns the provider-level default tags to merge for this network group,
// or nil when the resource opts out via ignore_default_tags.
func (r *ResourceNG) defaultTagsFor(ng Networkgroup) []string {
	if ng.IgnoreDefaultTags.ValueBool() {
		return nil
	}
	return r.DefaultTags()
}

// ModifyPlan recomputes the effective tags_all (provider default_tags merged with the
// resource tags, unless opted out). Network groups cannot be updated in place, so when the
// effective set changes on an existing resource, replacement is forced to apply the new tags.
func (r *ResourceNG) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.Provider == nil {
		return // resource is being destroyed, or provider not configured (e.g. validate)
	}

	plan := helper.PlanFrom[Networkgroup](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Can't compute the effective set until the resource tags are known.
	if plan.Tags.IsUnknown() {
		plan.TagsAll = basetypes.NewSetUnknown(types.StringType)
		resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
		return
	}

	merged := pkg.MergeTags(r.defaultTagsFor(plan), pkg.SetToStringSlice(ctx, plan.Tags, &resp.Diagnostics))
	plan.TagsAll = pkg.FromSetString(merged, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.State.Raw.IsNull() {
		return // create: nothing to replace
	}
	state := helper.StateFrom[Networkgroup](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Guard on a non-null prior tags_all so upgrading state that predates this attribute
	// (null tags_all) doesn't trigger a spurious replacement before the first refresh.
	if !state.TagsAll.IsNull() && !state.TagsAll.Equal(plan.TagsAll) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("tags_all"))
	}
}

// readFromAPI updates the given state with values returned by the API.
//
// tags_all reflects every tag on the network group. The resource-level `tags` attribute
// excludes the provider-level default_tags so they don't leak into it (which would cause
// a perpetual diff). For optional fields (description, tags), the rule is:
//   - state null + nothing resource-level → keep null (user never set it, nothing to sync)
//   - otherwise                           → sync from API
func readFromAPI(state *Networkgroup, ng *models.NetworkGroup1, defaultTags []string, diags *diag.Diagnostics) {
	if ng == nil || state == nil {
		return
	}

	state.Name = pkg.FromStrMaxLen(ng.Label)

	apiDescEmpty := ng.Description == nil || *ng.Description == ""
	if !state.Description.IsNull() || !apiDescEmpty {
		state.Description = basetypes.NewStringPointerValue(ng.Description)
	}

	// tags_all is the full effective set (resource tags + provider default_tags);
	// tags excludes the provider default_tags.
	state.Tags, state.TagsAll = pkg.SplitTags(ng.Tags, defaultTags, state.Tags, diags)

	state.Network = pkg.FromStr(ng.NetworkIP)
}

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
	// Apply the union of the provider-level default_tags and the resource-level tags
	// (unless the resource opts out via ignore_default_tags).
	defaultTags := r.defaultTagsFor(plan)
	mergedTags := pkg.MergeTags(defaultTags, pkg.SetToStringSlice(ctx, plan.Tags, &resp.Diagnostics))
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
			Tags:        mergedTags,
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

	readFromAPI(&plan, ng, defaultTags, &resp.Diagnostics)
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

	readFromAPI(&state, ng, r.defaultTagsFor(state), &resp.Diagnostics)
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
