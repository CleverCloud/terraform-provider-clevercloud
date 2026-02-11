package materiakv

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceMateriaKV) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	kv := helper.PlanFrom[MateriaKV](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	provider := pkg.LookupAddonProvider(*addonsProviders, "kv")

	plan := provider.FirstPlan()
	if plan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       kv.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "kv",
		Region:     kv.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	kv.ID = pkg.FromStr(res.Payload().RealID)
	kv.CreationDate = pkg.FromI(res.Payload().CreationDate)

	resp.Diagnostics.Append(resp.State.Set(ctx, kv)...)

	kvInfoRes := r.SDK.V4().Materia().
		Organisations().Ownerid(r.Organization()).Materia().
		Databases().Resourceid(kv.ID.ValueString()).Getmateriakvv4(ctx)
	if kvInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get materia kv connection infos", kvInfoRes.Error().Error())
		return
	}

	kvInfo := kvInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", kvInfo),
	})
	kv.Host = pkg.FromStr(kvInfo.Host)
	kv.Port = pkg.FromI(int64(kvInfo.Port))
	kv.Token = pkg.FromStr(kvInfo.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, kv)...)
}

// Read resource information
func (r *ResourceMateriaKV) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "MateriaKV READ", map[string]any{"request": req})

	var kv MateriaKV
	diags := req.State.Get(ctx, &kv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonKVRes := r.SDK.V4().Materia().
		Organisations().Ownerid(r.Organization()).Materia().
		Databases().Resourceid(kv.ID.ValueString()).Getmateriakvv4(ctx)
	if addonKVRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if addonKVRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonKVRes.HasError() {
		resp.Diagnostics.AddError("failed to get materiakv resource", addonKVRes.Error().Error())
	}

	addonKV := addonKVRes.Payload()

	if addonKV.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "STATE", map[string]any{"kv": kv})
	tflog.Debug(ctx, "API", map[string]any{"kv": addonKV})
	kv.Host = pkg.FromStr(addonKV.Host)
	kv.Port = pkg.FromI(int64(addonKV.Port))
	kv.Token = pkg.FromStr(addonKV.Token)

	diags = resp.State.Set(ctx, kv)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceMateriaKV) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[MateriaKV](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state := helper.StateFrom[MateriaKV](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update MateriaKV", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceMateriaKV) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	kv := helper.StateFrom[MateriaKV](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "MateriaKV DELETE", map[string]any{"kv": kv})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), kv.ID.ValueString())
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
