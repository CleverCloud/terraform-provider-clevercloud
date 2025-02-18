package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourcePostgreSQL) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourcePostgreSQL.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
	}

	r.Init(ctx, r.cc, r.org, "postgresql-addon", resp.Diagnostics)
}

// Create a new resource
func (r *ResourcePostgreSQL) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	pg := r.GetState(ctx, &req.Plan, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := r.LookupPlanBySlug(pg.Plan.ValueString())
	if plan.ID == "" {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(r.Plans()), ", ")+", got: "+pg.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       pg.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "postgresql-addon",
		Region:     pg.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	pg.ID = pkg.FromStr(res.Payload().ID)
	pg.CreationDate = pkg.FromI(res.Payload().CreationDate)
	pg.Plan = pkg.FromStr(res.Payload().Plan.Slug)

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
	tflog.Debug(ctx, "API response", map[string]interface{}{
		"payload": fmt.Sprintf("%+v", addonPG),
	})
	pg.Host = pkg.FromStr(addonPG.Host)
	pg.Port = pkg.FromI(int64(addonPG.Port))
	pg.Database = pkg.FromStr(addonPG.Database)
	pg.User = pkg.FromStr(addonPG.User)
	pg.Password = pkg.FromStr(addonPG.Password)

	r.SetState(ctx, &resp.State, *pg, resp.Diagnostics)
}

// Read resource information
func (r *ResourcePostgreSQL) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "PostgreSQL READ", map[string]interface{}{"request": req})

	pg := r.GetState(ctx, &req.State, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonPGRes := tmp.GetPostgreSQL(ctx, r.cc, pg.ID.ValueString())
	if addonPGRes.IsNotFoundError() {
		diags := resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
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

	tflog.Debug(ctx, "STATE", map[string]interface{}{"pg": pg})
	tflog.Debug(ctx, "API", map[string]interface{}{"pg": addonPG})
	pg.Plan = pkg.FromStr(addonPG.Plan)
	pg.Region = pkg.FromStr(addonPG.Zone)
	//pg.Name = types.String{Value: addonPG.}
	pg.Host = pkg.FromStr(addonPG.Host)
	pg.Port = pkg.FromI(int64(addonPG.Port))
	pg.Database = pkg.FromStr(addonPG.Database)
	pg.User = pkg.FromStr(addonPG.User)
	pg.Password = pkg.FromStr(addonPG.Password)

	r.SetState(ctx, &resp.State, *pg, resp.Diagnostics)
}

// Update resource
func (r *ResourcePostgreSQL) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourcePostgreSQL) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	pg := r.GetState(ctx, &req.State, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "PostgreSQL DELETE", map[string]interface{}{"pg": pg})

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
