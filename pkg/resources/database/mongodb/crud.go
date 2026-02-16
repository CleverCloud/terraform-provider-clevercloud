package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceMongoDB) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourceMongoDB.Create()")

	mg := helper.PlanFrom[MongoDB](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
	mg.Database = pkg.FromStr(addonMG.Database)
	mg.Uri = pkg.FromStr(addonMG.Uri())

	addon.SyncNetworkGroups(
		ctx,
		r,
		res.Payload().ID,
		mg.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, mg)...)
}

// Read resource information
func (r *ResourceMongoDB) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourceMongoDB.Read()")

	mg := helper.StateFrom[MongoDB](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), mg.ID.ValueString())
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
		resp.Diagnostics.AddError("failed to get MongoDB addon", addonRes.Error().Error())
		return
	}
	addonInfo := addonRes.Payload()
	mg.ID = pkg.FromStr(addonInfo.RealID)
	mg.Name = pkg.FromStr(addonInfo.Name)
	mg.Region = pkg.FromStr(addonInfo.Region)
	mg.Plan = pkg.FromStr(addonInfo.Plan.Slug)
	mg.CreationDate = pkg.FromI(addonInfo.CreationDate)

	addonMGRes := tmp.GetMongoDB(ctx, r.Client(), addonId)
	if addonMGRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonMGRes.HasError() {
		resp.Diagnostics.AddError("failed to get MongoDB resource", addonMGRes.Error().Error())
		return
	}
	addonMG := addonMGRes.Payload()

	if addonMG.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	mg.Host = pkg.FromStr(addonMG.Host)
	mg.Port = pkg.FromI(int64(addonMG.Port))
	mg.User = pkg.FromStr(addonMG.User)
	mg.Password = pkg.FromStr(addonMG.Password)
	mg.Database = pkg.FromStr(addonMG.Database)
	mg.Uri = pkg.FromStr(addonMG.Uri())

	mg.Networkgroups = resources.ReadNetworkGroups(ctx, r, addonId, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, mg)...)
}

// Update resource
func (r *ResourceMongoDB) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourceMongoDB.Update()")

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
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	addon.SyncNetworkGroups(
		ctx,
		r,
		plan.ID.ValueString(),
		plan.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceMongoDB) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	mg := helper.StateFrom[MongoDB](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ResourceMongoDB.Delete()")

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
