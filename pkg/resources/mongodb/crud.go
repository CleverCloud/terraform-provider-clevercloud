package mongodb

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
func (r *ResourceMongoDB) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceMongoDB.Configure()")

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
func (r *ResourceMongoDB) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	mg := MongoDB{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &mg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	prov := pkg.LookupAddonProvider(*addonsProviders, "mongodb-addon")
	plan := pkg.LookupProviderPlan(prov, mg.Plan.ValueString())
	if plan.ID == "" {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+mg.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       mg.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "mongodb-addon",
		Region:     mg.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	mg.ID = pkg.FromStr(res.Payload().ID)
	mg.CreationDate = pkg.FromI(res.Payload().CreationDate)
	mg.Plan = pkg.FromStr(res.Payload().Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, mg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mgInfoRes := tmp.GetMongoDB(ctx, r.cc, mg.ID.ValueString())
	if mgInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get MongoDB connection infos", mgInfoRes.Error().Error())
		return
	}

	addonMG := mgInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]interface{}{
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
	tflog.Debug(ctx, "MongoDB READ", map[string]interface{}{"request": req})

	var mg MongoDB
	diags := req.State.Get(ctx, &mg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonMGRes := tmp.GetMongoDB(ctx, r.cc, mg.ID.ValueString())
	if addonMGRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
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

	tflog.Debug(ctx, "STATE", map[string]interface{}{"mg": mg})
	tflog.Debug(ctx, "API", map[string]interface{}{"mg": addonMG})
	mg.Host = pkg.FromStr(addonMG.Host)
	mg.Port = pkg.FromI(int64(addonMG.Port))
	mg.User = pkg.FromStr(addonMG.User)
	mg.Password = pkg.FromStr(addonMG.Password)

	diags = resp.State.Set(ctx, mg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceMongoDB) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceMongoDB) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var mg MongoDB

	diags := req.State.Get(ctx, &mg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "MongoDB DELETE", map[string]interface{}{"mg": mg})

	res := tmp.DeleteAddon(ctx, r.cc, r.org, mg.ID.ValueString())
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
func (r *ResourceMongoDB) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
