package addon

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
func (r *ResourceAddon) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceAddon.Configure()")

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
func (r *ResourceAddon) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ad := Addon{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &ad)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()

	provider := pkg.LookupAddonProvider(*addonsProviders, ad.ThirdPartyProvider.ValueString())
	if provider == nil {
		resp.Diagnostics.AddError("This provider doesn't exist", fmt.Sprintf("available providers are: %s", strings.Join(pkg.AddonProvidersAsList(*addonsProviders), ", ")))
		return
	}

	plan := pkg.LookupProviderPlan(provider, ad.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("This plan doesn't exist", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       ad.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: provider.ID,
		Region:     ad.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}

	ad.ID = pkg.FromStr(res.Payload().ID)
	ad.CreationDate = pkg.FromI(res.Payload().CreationDate)

	envRes := tmp.GetAddonEnv(ctx, r.cc, r.org, res.Payload().ID)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", res.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*envRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceAddon) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Add-on READ", map[string]any{"request": req})

	var ad Addon
	diags := req.State.Get(ctx, &ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.cc, r.org, ad.ID.ValueString())
	if addonRes.IsNotFoundError() {
		req.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on", addonRes.Error().Error())
		return
	}

	addonEnvRes := tmp.GetAddonEnv(ctx, r.cc, r.org, ad.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*addonEnvRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})

	a := addonRes.Payload()
	ad.Name = pkg.FromStr(a.Name)
	ad.Plan = pkg.FromStr(a.Plan.Slug)
	ad.Region = pkg.FromStr(a.Region)
	ad.ThirdPartyProvider = pkg.FromStr(a.Provider.ID)
	ad.CreationDate = pkg.FromI(a.CreationDate)
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	diags = resp.State.Set(ctx, ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceAddon) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceAddon) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var ad Addon

	diags := req.State.Get(ctx, &ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Add-on DELETE", map[string]any{"addon": ad})

	res := tmp.DeleteAddon(ctx, r.cc, r.org, ad.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete add-on", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// Import resource
func (r *ResourceAddon) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
