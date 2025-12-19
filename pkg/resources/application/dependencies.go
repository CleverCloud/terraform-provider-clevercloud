package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// SyncDependencies synchronizes addon dependencies for an application.
// It compares expected dependencies (from Terraform plan) with current dependencies (from API)
// and adds/removes addons accordingly using Set.Difference() pattern.
func SyncDependencies(
	ctx context.Context,
	cc *client.Client,
	organization string,
	applicationID string,
	expectedDeps []string,
	diags *diag.Diagnostics,
) {
	// Get current linked addons from API
	currentAddonsRes := tmp.GetAppLinkedAddons(ctx, cc, organization, applicationID)
	if currentAddonsRes.HasError() {
		diags.AddError("failed to get linked addons", currentAddonsRes.Error().Error())
		return
	}

	// Extract addon IDs from current addons
	currentAddonIDs := pkg.Map(*currentAddonsRes.Payload(), func(addon tmp.AddonResponse) string {
		return addon.ID
	})

	// Create sets for comparison
	expectedSet := set.New(expectedDeps...)
	currentSet := set.New(currentAddonIDs...)

	tflog.Debug(ctx, "SYNC DEPENDENCIES", map[string]any{
		"expected": expectedDeps,
		"current":  currentAddonIDs,
	})

	// Remove dependencies that are no longer expected (current - expected)
	for addonID := range currentSet.Difference(expectedSet).Iter() {
		tflog.Info(ctx, "unlinking addon", map[string]any{"addonID": addonID})

		deleteRes := tmp.DeleteAppLinkedAddon(ctx, cc, organization, applicationID, addonID)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to unlink addon "+addonID, deleteRes.Error().Error())
		}
	}

	// Add new dependencies (expected - current)
	for addonID := range expectedSet.Difference(currentSet).Iter() {
		tflog.Info(ctx, "linking addon", map[string]any{"addonID": addonID})

		addRes := tmp.AddAppLinkedAddons(ctx, cc, organization, applicationID, addonID)
		if addRes.HasError() {
			diags.AddError("failed to link addon "+addonID, addRes.Error().Error())
		}
	}
}
