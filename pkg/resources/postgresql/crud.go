package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourcePostgreSQL) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourcePostgreSQL.Configure()")

	r.FetchPostgresInfos(ctx, &resp.Diagnostics)

	// Prevent panic if the provider has not been configured.
	if req.ProviderData != nil {

		provider, ok := req.ProviderData.(provider.Provider)
		if ok {
			r.cc = provider.Client()
			r.org = provider.Organization()
		}

		tflog.Debug(ctx, "AFTER CONFIGURED", map[string]any{"cc": r.cc == nil, "org": r.org})
	}
}

func (r *ResourcePostgreSQL) FetchPostgresInfos(ctx context.Context, diags *diag.Diagnostics) {
	cc := client.New()

	res := tmp.GetPostgresInfos(ctx, cc)
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
	pg := PostgreSQL{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &pg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	prov := pkg.LookupAddonProvider(*addonsProviders, "postgresql-addon")
	plan := pkg.LookupProviderPlan(prov, pg.Plan.ValueString())
	if plan.ID == "" {
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

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	createdPg := res.Payload()

	pg.ID = pkg.FromStr(createdPg.ID)
	pg.CreationDate = pkg.FromI(createdPg.CreationDate)
	pg.Plan = pkg.FromStr(createdPg.Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pgInfoRes := tmp.GetPostgreSQL(ctx, r.cc, pg.ID.ValueString())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourcePostgreSQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "PostgreSQL READ", map[string]any{"request": req})

	var pg PostgreSQL
	diags := req.State.Get(ctx, &pg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonPGRes := tmp.GetPostgreSQL(ctx, r.cc, pg.ID.ValueString())
	if addonPGRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if addonPGRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonPGRes.HasError() {
		resp.Diagnostics.AddError("failed to get Postgres resource", addonPGRes.Error().Error())
	}

	addonPG := addonPGRes.Payload()

	if addonPG.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "STATE", map[string]any{"pg": pg})
	tflog.Debug(ctx, "API", map[string]any{"pg": addonPG})
	pg.Plan = pkg.FromStr(addonPG.Plan)
	pg.Region = pkg.FromStr(addonPG.Zone)
	//pg.Name = types.String{Value: addonPG.}
	pg.Host = pkg.FromStr(addonPG.Host)
	pg.Port = pkg.FromI(int64(addonPG.Port))
	pg.Database = pkg.FromStr(addonPG.Database)
	pg.User = pkg.FromStr(addonPG.User)
	pg.Password = pkg.FromStr(addonPG.Password)
	pg.Version = pkg.FromStr(addonPG.Version)

	for _, feature := range addonPG.Features {
		if feature.Name == "do-backup" {
			pg.Backup = pkg.FromBool(feature.Enabled)
		}
	}

	diags = resp.State.Set(ctx, pg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourcePostgreSQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourcePostgreSQL) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var pg PostgreSQL

	diags := req.State.Get(ctx, &pg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "PostgreSQL DELETE", map[string]any{"pg": pg})

	res := tmp.DeleteAddon(ctx, r.cc, r.org, pg.ID.ValueString())
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
func (r *ResourcePostgreSQL) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
