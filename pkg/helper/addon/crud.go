package addon

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// GetAddonPlan fetches addon providers and looks up the specified plan for a provider.
// Returns the plan and diagnostics. If the plan is not found, adds an error diagnostic.
// Used by: MongoDB, MySQL, PostgreSQL, Redis, Generic
func GetAddonPlan(ctx context.Context, client *client.Client, providerID, planName string) (*tmp.AddonPlan, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, client)
	if addonsProvidersRes.HasError() {
		diags.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return nil, diags
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, providerID)
	if prov == nil {
		diags.AddError("provider not found", "provider ID: "+providerID)
		return nil, diags
	}

	plan := pkg.LookupProviderPlan(prov, planName)
	if plan == nil {
		diags.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", ")+", got: "+planName)
		return nil, diags
	}

	return plan, diags
}

// GetFirstAddonPlan fetches the first available plan for an addon provider.
// Used by addons that only have one plan (Keycloak, Matomo, Metabase, Cellar, FSBucket, MaterialKV, ConfigProvider, Otoroshi, Pulsar)
func GetFirstAddonPlan(ctx context.Context, client *client.Client, providerID string) (*tmp.AddonPlan, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, client)
	if addonsProvidersRes.HasError() {
		diags.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return nil, diags
	}
	addonsProviders := addonsProvidersRes.Payload()

	provider := pkg.LookupAddonProvider(*addonsProviders, providerID)
	if provider == nil {
		diags.AddError("provider not found", "provider ID: "+providerID)
		return nil, diags
	}

	plan := provider.FirstPlan()
	if plan == nil {
		diags.AddError("at least 1 plan for addon is required", "no plans available for provider: "+providerID)
		return nil, diags
	}

	return plan, diags
}

// Create handles the standard addon creation flow.
// Returns the created addon response and diagnostics.
// Used by most addon resources during Create operation.
func Create(ctx context.Context, client *client.Client, organization string, addonReq tmp.AddonRequest) (*tmp.AddonResponse, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	res := tmp.CreateAddon(ctx, client, organization, addonReq)
	if res.HasError() {
		diags.AddError("failed to create addon", res.Error().Error())
		return nil, diags
	}

	return res.Payload(), diags
}

// Read fetches an addon by ID and returns whether it was not found.
// Returns addon, isNotFound flag, and diagnostics.
// Used by addon resources during Read operation.
func Read(ctx context.Context, client *client.Client, organization, addonID string) (*tmp.AddonResponse, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	addonRes := tmp.GetAddon(ctx, client, organization, addonID)
	if addonRes.IsNotFoundError() {
		return nil, true, diags
	}
	if addonRes.HasError() {
		diags.AddError("failed to get addon", addonRes.Error().Error())
		return nil, false, diags
	}

	return addonRes.Payload(), false, diags
}

// Update updates addon properties.
// Accepts a map of fields to update (e.g., {"name": "new-name", "plan": "new-plan"}).
// Currently used by addon resources to update the name field.
func Update(ctx context.Context, client *client.Client, organization, addonID string, fields map[string]string) diag.Diagnostics {
	diags := diag.Diagnostics{}

	addonRes := tmp.UpdateAddon(ctx, client, organization, addonID, fields)
	if addonRes.HasError() {
		diags.AddError("failed to update addon", addonRes.Error().Error())
	}

	return diags
}

// Delete deletes an addon and handles the NotFoundError case consistently.
// Returns diagnostics on error.
// Used by all addon resources during Delete operation.
func Delete(ctx context.Context, client *client.Client, organization, addonID string) diag.Diagnostics {
	diags := diag.Diagnostics{}

	res := tmp.DeleteAddon(ctx, client, organization, addonID)
	if res.IsNotFoundError() {
		// Addon already deleted, this is fine
		return diags
	}
	if res.HasError() {
		diags.AddError("failed to delete addon", res.Error().Error())
	}

	return diags
}
