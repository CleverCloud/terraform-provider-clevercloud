package postgresql

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

func (r *ResourcePostgreSQL) FetchPostgresInfos(ctx context.Context, diags *diag.Diagnostics) {
	// Skip fetching during schema validation (before provider is configured)
	if r.Provider == nil || r.Client() == nil {
		tflog.Debug(ctx, "Skipping postgres infos fetch - provider not configured yet")
		return
	}

	res := tmp.GetPostgresInfos(ctx, r.Client())
	if res.HasError() {
		tflog.Error(ctx, "failed to get postgres infos", map[string]any{"error": res.Error().Error()})
		return
	}
	r.infos = res.Payload()
	for k := range r.infos.Dedicated {
		r.dedicatedVersions = append(r.dedicatedVersions, k)
	}
}

func (r *ResourcePostgreSQL) Infos(ctx context.Context, diags *diag.Diagnostics) *tmp.PostgresInfos {
	if r.infos == nil {
		r.FetchPostgresInfos(ctx, diags)
	}

	return r.infos
}

// Create a new resource
func (r *ResourcePostgreSQL) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	pg := helper.PlanFrom[PostgreSQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "postgresql-addon")
	plan := pkg.LookupProviderPlan(prov, pg.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+pg.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       pg.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "postgresql-addon",
		Region:     pg.Region.ValueString(),
		Options:    map[string]string{},
	}

	if !pg.Version.IsNull() && !pg.Version.IsUnknown() {
		addonReq.Options["version"] = pg.Version.ValueString()
	}

	if !pg.Backup.IsNull() && !pg.Backup.IsUnknown() {
		backupValue := fmt.Sprintf("%t", pg.Backup.ValueBool())
		addonReq.Options["do-backup"] = backupValue
	} else {
		addonReq.Options["do-backup"] = "true"
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	createdPg := res.Payload()

	pg.ID = pkg.FromStr(createdPg.RealID)
	pg.CreationDate = pkg.FromI(createdPg.CreationDate)
	pg.Plan = pkg.FromStr(createdPg.Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)

	pgInfoRes := tmp.GetPostgreSQL(ctx, r.Client(), createdPg.ID)
	if pgInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get postgres connection infos", pgInfoRes.Error().Error())
		return
	}
	addonPG := pgInfoRes.Payload()

	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", addonPG),
	})
	pg.Host = pkg.FromStr(addonPG.Host)
	pg.Port = pkg.FromI(int64(addonPG.Port))
	pg.Database = pkg.FromStr(addonPG.Database)
	pg.User = pkg.FromStr(addonPG.User)
	pg.Password = pkg.FromStr(addonPG.Password)
	pg.Version = pkg.FromStr(addonPG.Version)
	pg.Uri = pkg.FromStr(addonPG.Uri())

	addon.SyncNetworkGroups(
		ctx,
		r.Client(),
		r.Organization(),
		createdPg.ID,
		pg.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
}

// Read resource information
func (r *ResourcePostgreSQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "PostgreSQL READ", map[string]any{"request": req})

	// State
	pg := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// IDs
	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), pg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	realID, err := tmp.AddonIDToRealID(ctx, r.Client(), r.Organization(), pg.ID.ValueString())
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
		resp.Diagnostics.AddError("failed to get Postgres resource", addonRes.Error().Error())
		return
	}
	addonInfo := addonRes.Payload()

	addonPGRes := tmp.GetPostgreSQL(ctx, r.Client(), addonId)
	if addonPGRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonPGRes.HasError() {
		resp.Diagnostics.AddError("failed to get Postgres resource", addonPGRes.Error().Error())
		return
	}

	addonPG := addonPGRes.Payload()

	if addonPG.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "STATE", map[string]any{"pg": pg})
	tflog.Debug(ctx, "API", map[string]any{"pg": addonPG})
	pg.ID = pkg.FromStr(realID)
	pg.Name = pkg.FromStr(addonInfo.Name)
	pg.Plan = pkg.FromStr(addonPG.Plan)
	pg.Region = pkg.FromStr(addonPG.Zone)
	pg.CreationDate = pkg.FromI(addonInfo.CreationDate)
	pg.Host = pkg.FromStr(addonPG.Host)
	pg.Port = pkg.FromI(int64(addonPG.Port))
	pg.Database = pkg.FromStr(addonPG.Database)
	pg.User = pkg.FromStr(addonPG.User)
	pg.Password = pkg.FromStr(addonPG.Password)
	pg.Version = pkg.FromStr(addonPG.Version)
	pg.Uri = pkg.FromStr(addonPG.Uri())

	for _, feature := range addonPG.Features {
		if feature.Name == "do-backup" {
			pg.Backup = pkg.FromBool(feature.Enabled)
		}
	}

	pg.Networkgroups = resources.ReadNetworkGroups(ctx, r.Client(), r.Organization(), addonId, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
}

// Update resource
func (r *ResourcePostgreSQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[PostgreSQL](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("postgresql cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update PostgreSQL", addonRes.Error().Error())
		return
	}
	state.Name = plan.Name

	addon.SyncNetworkGroups(
		ctx,
		r.Client(),
		r.Organization(),
		plan.ID.ValueString(),
		plan.Networkgroups,
		&resp.Diagnostics,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourcePostgreSQL) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	pg := helper.StateFrom[PostgreSQL](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "PostgreSQL DELETE", map[string]any{"pg": pg})

	addonId, err := tmp.RealIDToAddonID(ctx, r.Client(), r.Organization(), pg.ID.ValueString())
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
