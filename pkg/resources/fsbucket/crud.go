package fsbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceFSBucket) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceFSBucket.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}
}

// Create a new resource
func (r *ResourceFSBucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fsbucket := helper.PlanFrom[FSBucket](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
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

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
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
	envRes := tmp.GetAddonEnv(ctx, r.cc, r.org, addonRes.RealID)
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
	tflog.Debug(ctx, "FSBucket READ", map[string]any{"request": req})

	fsbucket := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.cc, r.org, fsbucket.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get FSBucket", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()

	addonEnvRes := tmp.GetAddonEnv(ctx, r.cc, r.org, fsbucket.ID.ValueString())
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
	addonRes := tmp.UpdateAddon(ctx, r.cc, r.org, plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update FSBucket", addonRes.Error().Error())
		return
	}
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceFSBucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	fsbucket := helper.StateFrom[FSBucket](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "FSBUCKET DELETE", map[string]any{"fsbucket": fsbucket})

	addonRes := tmp.GetAddon(ctx, r.cc, r.org, fsbucket.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Addon", addonRes.Error().Error())
		return
	}

	res := tmp.DeleteAddon(ctx, r.cc, r.org, fsbucket.ID.ValueString())
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

// Import resource
func (r *ResourceFSBucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
