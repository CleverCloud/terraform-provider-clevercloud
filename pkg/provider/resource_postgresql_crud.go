package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourcePostgreSQL) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "ResourcePostgreSQL.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*Provider)
	if ok {
		r.cc = provider.cc
		r.org = provider.Organisation
	}
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

	var plan tmp.AddonPlan
	addonsProviders := addonsProvidersRes.Payload()
	for i := range *addonsProviders {
		addonsProvider := (*addonsProviders)[i]
		if addonsProvider.ID == "postgresql-addon" {
			for _, pl := range addonsProvider.Plans {
				if pl.Slug == pg.Plan.ValueString() {
					tflog.Info(ctx, "Plan matched", map[string]interface{}{"name": pg.Plan.ValueString(), "plan": pl.Slug})
					plan = pl
				}
			}
		}
	}
	if plan.ID == "" {
		resp.Diagnostics.AddError("failed to find plan", "expect:, got: "+pg.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       pg.Name.String(),
		Plan:       plan.ID,
		ProviderID: "postgresql-addon",
		Region:     pg.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	pg.ID = fromStr(res.Payload().ID)
	pg.CreationDate = fromI(res.Payload().CreationDate)
	pg.Plan = fromStr(res.Payload().Plan.Slug)

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
	tflog.Info(ctx, "API response", map[string]interface{}{
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
func (r *ResourcePostgreSQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "PostgreSQL READ", map[string]interface{}{"request": req})

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
	if addonPGRes.HasError() {
		resp.Diagnostics.AddError("failed to get Postgres resource", addonPGRes.Error().Error())
	}

	addonPG := addonPGRes.Payload()
	tflog.Info(ctx, "STATE", map[string]interface{}{"pg": pg})
	tflog.Info(ctx, "API", map[string]interface{}{"pg": addonPG})
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
	tflog.Info(ctx, "PostgreSQL DELETE", map[string]interface{}{"pg": pg})

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
