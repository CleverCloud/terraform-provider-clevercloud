package metabase

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceMetabase) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceMetabase.Configure()")

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
func (r *ResourceMetabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	mb := helper.PlanFrom[Metabase](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	prov := pkg.LookupAddonProvider(*addonsProviders, "metabase")
	plan := pkg.LookupProviderPlan(prov, mb.Plan.ValueString())
	if plan == nil || plan.ID == "" {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+mb.Plan.String())
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       mb.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "metabase",
		Region:     mb.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}

	mb.ID = pkg.FromStr(res.Payload().RealID)
	mb.CreationDate = pkg.FromI(res.Payload().CreationDate)
	// mb.Plan = pkg.FromStr(res.Payload().Plan.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, mb)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mbInfoRes := tmp.GetAddonEnv(ctx, r.cc, r.org, res.Payload().ID)
	if mbInfoRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase connection infos", mbInfoRes.Error().Error())
		return
	}

	addonMB := *mbInfoRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", addonMB),
	})

	hostEnvVar := pkg.First(addonMB, func(v tmp.EnvVar) bool {
		return v.Name == "METABASE_URL"
	})
	if hostEnvVar == nil {
		resp.Diagnostics.AddError("cannot get Metabase infos", "missing METABASE_URL env var on created addon")
		return
	}

	mb.Host = pkg.FromStr(hostEnvVar.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, mb)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceMetabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Metabase READ", map[string]any{"request": req})

	mb := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonMBRes := tmp.GetMetabase(ctx, r.cc, mb.ID.ValueString())
	if addonMBRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonMBRes.HasError() {
		resp.Diagnostics.AddError("failed to get Metabase resource", addonMBRes.Error().Error())
		return
	}

	addonMB := addonMBRes.Payload()

	if addonMB.Status == "TO_DELETE" {
		resp.State.RemoveResource(ctx)
		return
	}

	realID, err := tmp.AddonIDToRealID(ctx, r.cc, r.org, mb.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	tflog.Debug(ctx, "STATE", map[string]any{"mb": mb})
	tflog.Debug(ctx, "API", map[string]any{"mb": addonMB})
	mb.ID = pkg.FromStr(realID)
	// mb.Host = pkg.FromStr(addonMB.Applications[0].Host)
	// mb.Port = pkg.FromI(int64(addonMB.Port))
	// mb.User = pkg.FromStr(addonMB.User)
	// mb.Password = pkg.FromStr(addonMB.Password)

	diags := resp.State.Set(ctx, mb)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceMetabase) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceMetabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	mb := helper.StateFrom[Metabase](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Metabase DELETE", map[string]any{"mb": mb})

	addonId, err := tmp.RealIDToAddonID(ctx, r.cc, r.org, mb.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get addon ID", err.Error())
		return
	}

	res := tmp.DeleteAddon(ctx, r.cc, r.org, addonId)
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
func (r *ResourceMetabase) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
