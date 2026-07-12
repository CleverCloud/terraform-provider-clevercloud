package fsbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// ModifyPlan recomputes the effective `tags_all` (provider defaults merged with the
// resource `tags`) at plan time, so a provider-level default_tags change propagates to
// existing resources.
func (r *ResourceFSBucket) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.Provider == nil {
		return
	}

	plan := helper.PlanFrom[FSBucket](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.TagsAll = pkg.ComputeTagsAll(ctx, r.DefaultTags(), plan.Tags, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

// Create a new resource
func (r *ResourceFSBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fsbucket := helper.PlanFrom[FSBucket](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "fs-bucket")
	if prov == nil {
		resp.Diagnostics.AddError("failed to find fs-bucket provider", "")
		return
	}
	plan := prov.Plans[0]

	addonReq := tmp.AddonRequest{
		Name:       fsbucket.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "fs-bucket",
		Region:     fsbucket.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	addonRes := res.Payload()

	fsbucket.ID = pkg.FromStr(addonRes.RealID)
	fsbucket.Name = pkg.FromStr(addonRes.Name)
	fsbucket.Region = pkg.FromStr(addonRes.Region)

	resp.Diagnostics.Append(resp.State.Set(ctx, fsbucket)...)

	tflog.Debug(ctx, "get addon env vars", map[string]any{"fsbucket": addonRes.RealID})
	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), addonRes.RealID)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon env vars", envRes.Error().Error())
		return
	}

	envVars := envRes.Payload()
	envMap := pkg.Reduce(*envVars, map[string]types.String{}, func(m map[string]types.String, v tmp.EnvVar) map[string]types.String {
		m[v.Name] = pkg.FromStr(v.Value)
		return m
	})

	fsbucket.Host = envMap["BUCKET_HOST"]
	fsbucket.FTPUsername = envMap["BUCKET_FTP_USERNAME"]
	fsbucket.FTPPassword = envMap["BUCKET_FTP_PASSWORD"]

	resources.SyncTags(ctx, r, resources.AddonTags, addonRes.ID, fsbucket.Tags, &fsbucket.Tags, &fsbucket.TagsAll, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, fsbucket)...)
}

// Read resource information
func (r *ResourceFSBucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "FSBucket READ", map[string]any{"request": req})

	fsbucket := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if fsbucket.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), fsbucket.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get FSBucket", addonRes.Error().Error())
		return
	}
	addonPayload := addonRes.Payload()

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), fsbucket.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon env", addonEnvRes.Error().Error())
		return
	}
	addonEnv := addonEnvRes.Payload()
	addonMap := pkg.Reduce(*addonEnv, map[string]types.String{}, func(m map[string]types.String, v tmp.EnvVar) map[string]types.String {
		m[v.Name] = pkg.FromStr(v.Value)
		return m
	})

	fsbucket.Name = pkg.FromStr(addonPayload.Name)
	fsbucket.Region = pkg.FromStr(addonPayload.Region)
	fsbucket.Host = addonMap["BUCKET_HOST"]
	fsbucket.FTPUsername = addonMap["BUCKET_FTP_USERNAME"]
	fsbucket.FTPPassword = addonMap["BUCKET_FTP_PASSWORD"]

	fsbucket.Tags, fsbucket.TagsAll = resources.ReadTags(ctx, r, resources.AddonTags, addonPayload.ID, fsbucket.Tags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, fsbucket)...)
}

// Update resource
func (r *ResourceFSBucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[FSBucket](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID != state.ID {
		resp.Diagnostics.AddError("fsbucket cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update FSBucket", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resources.SyncTags(ctx, r, resources.AddonTags, addonRes.Payload().ID, plan.Tags, &state.Tags, &state.TagsAll, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceFSBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fsbucket := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "FSBUCKET DELETE", map[string]any{"fsbucket": fsbucket})

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), fsbucket.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Addon", addonRes.Error().Error())
		return
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), fsbucket.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete addon", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
