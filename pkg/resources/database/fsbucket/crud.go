package fsbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceFSBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceFSBucket.Create()")

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

	plan := prov.FirstPlan()
	if plan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

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

	resp.Diagnostics.Append(resp.State.Set(ctx, fsbucket)...)
}

// Read resource information
func (r *ResourceFSBucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceFSBucket.Read()")

	fsbucket := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), fsbucket.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonId)
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get FSBucket", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), addonId)
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon env", addonEnvRes.Error().Error())
		return
	}
	addonEnv := addonEnvRes.Payload()
	addonMap := pkg.Reduce(*addonEnv, map[string]types.String{}, func(m map[string]types.String, v tmp.EnvVar) map[string]types.String {
		m[v.Name] = pkg.FromStr(v.Value)
		return m
	})

	fsbucket.Name = pkg.FromStr(addon.Name)
	fsbucket.Region = pkg.FromStr(addon.Region)
	fsbucket.Host = addonMap["BUCKET_HOST"]
	fsbucket.FTPUsername = addonMap["BUCKET_FTP_USERNAME"]
	fsbucket.FTPPassword = addonMap["BUCKET_FTP_PASSWORD"]

	resp.Diagnostics.Append(resp.State.Set(ctx, fsbucket)...)
}

// Update resource
func (r *ResourceFSBucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceFSBucket.Update()")

	plan := helper.PlanFrom[FSBucket](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("fsbucket cannot be updated", "mismatched IDs")
		return
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), addonId, map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update FSBucket", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceFSBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fsbucket := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "ResourceFSBucket.Delete()")

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), fsbucket.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonId)
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
