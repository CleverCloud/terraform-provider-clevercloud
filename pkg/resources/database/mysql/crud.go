package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceMySQL) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.FetchMysqlInfos(ctx, &resp.Diagnostics)
	r.Configurer.Configure(ctx, req, resp)
}

func (r *ResourceMySQL) FetchMysqlInfos(ctx context.Context, diags *diag.Diagnostics) {
	// Skip fetching during schema validation (before provider is configured)
	if r.Provider == nil || r.Client() == nil {
		tflog.Debug(ctx, "Skipping mysql infos fetch - provider not configured yet")
		return
	}

	res := tmp.GetMysqlInfos(ctx, r.Client())
	if res.HasError() {
		tflog.Error(ctx, "failed to get mysql infos", map[string]any{"error": res.Error().Error()})
		return
	}
	r.infos = res.Payload()
	for k := range r.infos.Dedicated {
		r.dedicatedVersions = append(r.dedicatedVersions, k)
	}
}

func (r *ResourceMySQL) Infos(ctx context.Context, diags *diag.Diagnostics) *tmp.MysqlInfos {
	if r.infos == nil {
		r.FetchMysqlInfos(ctx, diags)
	}

	return r.infos
}

// Create a new resource
func (r *ResourceMySQL) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	my := helper.PlanFrom[MySQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "mysql-addon")
	plan := pkg.LookupProviderPlan(prov, my.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+my.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       my.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "mysql-addon",
		Region:     my.Region.ValueString(),
		Options:    map[string]string{},
	}

	if !my.Version.IsNull() && !my.Version.IsUnknown() {
		addonReq.Options["version"] = my.Version.ValueString()
	}

	addonReq.Options["do-backup"] = "true"
	if !my.Backup.IsNull() && !my.Backup.IsUnknown() {
		addonReq.Options["do-backup"] = fmt.Sprintf("%t", my.Backup.ValueBool())
	}

	if !my.Encryption.IsNull() && !my.Encryption.IsUnknown() {
		addonReq.Options["encryption"] = fmt.Sprintf("%t", my.Encryption.ValueBool())
	}

	if !my.DirectHostOnly.IsNull() && !my.DirectHostOnly.IsUnknown() {
		addonReq.Options["direct-host-only"] = fmt.Sprintf("%t", my.DirectHostOnly.ValueBool())
	}

	if !my.SkipLogBin.IsNull() && !my.SkipLogBin.IsUnknown() {
		addonReq.Options["skip-log-bin"] = fmt.Sprintf("%t", my.SkipLogBin.ValueBool())
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	createdMy := res.Payload()

	my.ID = pkg.FromStr(createdMy.RealID)
	my.CreationDate = pkg.FromI(createdMy.CreationDate)
	my.Plan = pkg.FromStr(createdMy.Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, my)...)

	myInfoRes := tmp.GetMySQL(ctx, r.Client(), createdMy.ID)
	if myInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get mysql connection infos", myInfoRes.Error().Error())
		return
	}
	addonMy := myInfoRes.Payload()

	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", addonMy),
	})
	my.Host = pkg.FromStr(addonMy.Host)
	my.Port = pkg.FromI(int64(addonMy.Port))
	my.Database = pkg.FromStr(addonMy.Database)
	my.User = pkg.FromStr(addonMy.User)
	my.Password = pkg.FromStr(addonMy.Password)
	my.Version = pkg.FromStr(addonMy.Version)
	my.Uri = pkg.FromStr(addonMy.Uri())
	my.ReadOnlyUsers = tmp.FromMySQLReadOnlyUsers(addonMy.ReadOnlyUsers)

	if my.Encryption.IsUnknown() {
		my.Encryption = pkg.FromBool(false)
	}
	if my.DirectHostOnly.IsUnknown() {
		my.DirectHostOnly = pkg.FromBool(false)
	}
	if my.SkipLogBin.IsUnknown() {
		my.SkipLogBin = pkg.FromBool(false)
	}
	for _, feature := range addonMy.Features {
		switch feature.Name {
		case "do-backup":
			my.Backup = pkg.FromBool(feature.Enabled)
		case "encryption":
			my.Encryption = pkg.FromBool(feature.Enabled)
		case "direct-host-only":
			my.DirectHostOnly = pkg.FromBool(feature.Enabled)
		case "skip-log-bin":
			my.SkipLogBin = pkg.FromBool(feature.Enabled)
		}
	}

	addon.SyncNetworkGroups(
		ctx,
		r,
		createdMy.ID,
		my.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, my)...)
}

// Read resource information
func (r *ResourceMySQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Mysql READ", map[string]any{"request": req})

	// State
	my := helper.StateFrom[MySQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// IDs
	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), my.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	// Objects
	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), addonId)
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Mysql resource", addonRes.Error().Error())
		return
	}
	addonInfo := addonRes.Payload()

	addonMyRes := tmp.GetMySQL(ctx, r.Client(), addonId)
	if addonMyRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonMyRes.HasError() {
		resp.Diagnostics.AddError("failed to get Mysql resource", addonMyRes.Error().Error())
		return
	}

	addonMy := addonMyRes.Payload()

	if addonMy.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "STATE", map[string]any{"my": my})
	tflog.Debug(ctx, "API", map[string]any{"my": addonMy})
	my.ID = pkg.FromStr(addonInfo.RealID)
	my.Name = pkg.FromStr(addonInfo.Name)
	my.Plan = pkg.FromStr(addonMy.Plan)
	my.Region = pkg.FromStr(addonMy.Zone)
	my.CreationDate = pkg.FromI(addonInfo.CreationDate)
	my.Host = pkg.FromStr(addonMy.Host)
	my.Port = pkg.FromI(int64(addonMy.Port))
	my.Database = pkg.FromStr(addonMy.Database)
	my.User = pkg.FromStr(addonMy.User)
	my.Password = pkg.FromStr(addonMy.Password)
	my.Version = pkg.FromStr(addonMy.Version)
	my.Uri = pkg.FromStr(addonMy.Uri())
	my.ReadOnlyUsers = tmp.FromMySQLReadOnlyUsers(addonMy.ReadOnlyUsers)

	for _, feature := range addonMy.Features {
		switch feature.Name {
		case "do-backup":
			my.Backup = pkg.FromBool(feature.Enabled)
		case "encryption":
			my.Encryption = pkg.FromBool(feature.Enabled)
		case "direct-host-only":
			my.DirectHostOnly = pkg.FromBool(feature.Enabled)
		case "skip-log-bin":
			my.SkipLogBin = pkg.FromBool(feature.Enabled)
		}
	}

	my.Networkgroups = resources.ReadNetworkGroups(ctx, r, addonId, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, my)...)
}

// Update resource
func (r *ResourceMySQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[MySQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[MySQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("mysql cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Mysql", addonRes.Error().Error())
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
func (r *ResourceMySQL) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	my := helper.StateFrom[MySQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "MySQL DELETE", map[string]any{"my": my})

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), my.ID.ValueString())
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
