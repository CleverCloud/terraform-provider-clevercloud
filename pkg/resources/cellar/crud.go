package cellar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceCellar) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "ResourceCellar.Configure()")

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
func (r *ResourceCellar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cellar := Cellar{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.cc)
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "cellar-addon")
	if prov == nil {
		resp.Diagnostics.AddError("failed to fin provider", "")
		return
	}
	plan := prov.Plans[0]

	addonReq := tmp.AddonRequest{
		Name:       cellar.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "cellar-addon",
		Region:     cellar.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.cc, r.org, addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon", res.Error().Error())
		return
	}
	addonRes := res.Payload()

	tflog.Debug(ctx, "get addon env vars", map[string]any{"cellar": addonRes.RealID})
	envRes := tmp.GetAddonEnv(ctx, r.cc, r.org, addonRes.RealID)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon env vars", envRes.Error().Error())
		return
	}

	envVars := envRes.Payload()
	creds := s3.FromEnvVars(*envVars)

	cellar.ID = pkg.FromStr(addonRes.RealID)
	cellar.Host = pkg.FromStr(creds.Host)
	cellar.KeyID = pkg.FromStr(creds.KeyID)
	cellar.KeySecret = pkg.FromStr(creds.KeySecret)

	resp.Diagnostics.Append(resp.State.Set(ctx, cellar)...)
}

// Read resource information
func (r *ResourceCellar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Cellar READ", map[string]any{"request": req})

	var cellar Cellar
	resp.Diagnostics.Append(req.State.Get(ctx, &cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO

	resp.Diagnostics.Append(resp.State.Set(ctx, cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r *ResourceCellar) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceCellar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var cellar Cellar

	resp.Diagnostics.Append(req.State.Get(ctx, &cellar)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "CELLAR DELETE", map[string]any{"cellar": cellar})

	addonRes := tmp.GetAddon(ctx, r.cc, r.org, cellar.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Addon", addonRes.Error().Error())
		return
	}

	// TODO: Use real ID when API support it
	// res := tmp.DeleteAddon(ctx, r.cc, r.org, cellar.ID.ValueString())
	res := tmp.DeleteAddon(ctx, r.cc, r.org, addonRes.Payload().ID)
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
func (r *ResourceCellar) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
