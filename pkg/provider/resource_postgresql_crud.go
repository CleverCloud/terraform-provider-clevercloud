package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ResourcePostgreSQL struct {
	cc  *client.Client
	org string
}

// Create a new resource
func (r ResourcePostgreSQL) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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

	var plan tmp.AddonPlan
	addonsProviders := addonsProvidersRes.Payload()
	for i := range *addonsProviders {
		addonsProvider := (*addonsProviders)[i]
		if addonsProvider.ID == "postgresql-addon" {
			for _, pl := range addonsProvider.Plans {
				if pl.Slug == pg.Plan.Value {
					tflog.Info(ctx, "Plan matched", map[string]interface{}{"name": pg.Plan.Value, "plan": pl.Slug})
					plan = pl
				}
			}
		}
	}
	if plan.ID == "" {
		resp.Diagnostics.AddError("failed to find plan", "expect:, got: "+pg.Plan.Value)
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       pg.Name.Value,
		Plan:       plan.ID,
		ProviderID: "postgresql-addon",
		Region:     pg.Region.Value,
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	pg.ID = fromStr(res.Payload().ID)
	pg.CreationDate = fromI(res.Payload().CreationDate)
	pg.Plan = fromStr(res.Payload().Plan.Slug)
	tflog.Info(ctx, "create response", map[string]interface{}{"plan": res.Payload()})

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pgInfoRes := tmp.GetPostgreSQL(ctx, r.cc, pg.ID.Value)
	if pgInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get postgres connection infos", pgInfoRes.Error().Error())
		return
	}

	addonPG := pgInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]interface{}{
		"payload": fmt.Sprintf("%+v", addonPG),
	})
	pg.Host = fromStr(addonPG.Host)
	pg.Port = fromI(int64(addonPG.Port))
	pg.Database = fromStr(addonPG.Database)
	pg.User = fromStr(addonPG.User)
	pg.Password = fromStr(addonPG.Password)

	resp.Diagnostics.Append(resp.State.Set(ctx, pg)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r ResourcePostgreSQL) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	tflog.Debug(ctx, "PostgreSQL READ", map[string]interface{}{"request": req})

	var pg PostgreSQL
	diags := req.State.Get(ctx, &pg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonPGRes := tmp.GetPostgreSQL(ctx, r.cc, pg.ID.Value)
	if addonPGRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), types.String{Unknown: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if addonPGRes.HasError() {
		resp.Diagnostics.AddError("failed to get Postgres resource", addonPGRes.Error().Error())
	}

	addonPG := addonPGRes.Payload()
	tflog.Debug(ctx, "STATE", map[string]interface{}{"pg": pg})
	tflog.Debug(ctx, "API", map[string]interface{}{"pg": addonPG})
	pg.Plan = fromStr(addonPG.Plan)
	pg.Region = fromStr(addonPG.Zone)
	//pg.Name = types.String{Value: addonPG.}
	pg.Host = fromStr(addonPG.Host)
	pg.Port = fromI(int64(addonPG.Port))
	pg.Database = fromStr(addonPG.Database)
	pg.User = fromStr(addonPG.User)
	pg.Password = fromStr(addonPG.Password)

	diags = resp.State.Set(ctx, pg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r ResourcePostgreSQL) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// TODO
}

// Delete resource
func (r ResourcePostgreSQL) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var pg PostgreSQL

	diags := req.State.Get(ctx, &pg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "PostgreSQL DELETE", map[string]interface{}{"pg": pg})

	res := tmp.DeletePostgres(ctx, r.cc, r.org, pg.ID.Value)
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
func (r ResourcePostgreSQL) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
