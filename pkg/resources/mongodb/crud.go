package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceMongoDB) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	mg := helper.PlanFrom[MongoDB](ctx, req.Plan, &resp.Diagnostics)

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "mongodb-addon")
	plan := pkg.LookupProviderPlan(prov, mg.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+mg.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       mg.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "mongodb-addon",
		Region:     mg.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	mg.ID = pkg.FromStr(res.Payload().RealID)
	mg.CreationDate = pkg.FromI(res.Payload().CreationDate)
	mg.Plan = pkg.FromStr(res.Payload().Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, mg)...)

	mgInfoRes := tmp.GetMongoDB(ctx, r.Client(), res.Payload().ID)
	if mgInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get MongoDB connection infos", mgInfoRes.Error().Error())
		return
	}

	addonMG := mgInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", addonMG),
	})
	mg.Host = pkg.FromStr(addonMG.Host)
	mg.Port = pkg.FromI(int64(addonMG.Port))
	mg.User = pkg.FromStr(addonMG.User)
	mg.Password = pkg.FromStr(addonMG.Password)

	resp.Diagnostics.Append(resp.State.Set(ctx, mg)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceMongoDB) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "MongoDB READ", map[string]any{"request": req})

	mg := helper.StateFrom[MongoDB](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), mg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	addonMGRes := tmp.GetMongoDB(ctx, r.Client(), addonId)
	if addonMGRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonMGRes.HasError() {
		resp.Diagnostics.AddError("failed to get MongoDB resource", addonMGRes.Error().Error())
	}

	addonMG := addonMGRes.Payload()

	if addonMG.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	realID, err := tmp.AddonIDToRealID(ctx, r.Client(), r.Organization(), mg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	tflog.Debug(ctx, "STATE", map[string]any{"mg": mg})
	tflog.Debug(ctx, "API", map[string]any{"mg": addonMG})
	mg.ID = pkg.FromStr(realID)
	mg.Host = pkg.FromStr(addonMG.Host)
	mg.Port = pkg.FromI(int64(addonMG.Port))
	mg.User = pkg.FromStr(addonMG.User)
	mg.Password = pkg.FromStr(addonMG.Password)

	diags := resp.State.Set(ctx, mg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceMongoDB) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[MongoDB](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[MongoDB](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("mongodb cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update MongoDB", addonRes.Error().Error())
		return
	}
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourceMongoDB) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	mg := helper.StateFrom[MongoDB](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "MongoDB DELETE", map[string]any{"mg": mg})

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), mg.ID.ValueString())
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

// Import resource
