package application

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// splitDependencies separates app IDs from addon IDs in a list of dependencies.
// App IDs have the format "app_xxx" while addon dependencies are everything else
// (addon_xxx, postgresql_xxx, mysql_xxx, etc.)
func splitDependencies(deps []string) (appIDs, addonIDs []string) {
	for _, dep := range deps {
		if strings.HasPrefix(dep, "app_") {
			appIDs = append(appIDs, dep)
		} else {
			addonIDs = append(addonIDs, dep)
		}
	}
	return
}

// ReadDependencies reads all dependencies (apps + addons) from API and returns them as a Set.
// stateValue is used to preserve null vs empty semantics.
func ReadDependencies(ctx context.Context, cc *client.Client, organization, applicationID string, stateValue types.Set, diags *diag.Diagnostics) types.Set {
	var allDeps []string

	// Read app dependencies
	appsRes := tmp.GetAppDependencies(ctx, cc, organization, applicationID)
	if appsRes.HasError() {
		diags.AddError("failed to get linked apps", appsRes.Error().Error())
		return types.SetNull(types.StringType)
	}
	for _, app := range *appsRes.Payload() {
		allDeps = append(allDeps, app.ID)
	}

	// Read addon dependencies
	addonsRes := tmp.GetAppLinkedAddons(ctx, cc, organization, applicationID)
	if addonsRes.HasError() {
		diags.AddError("failed to get linked addons", addonsRes.Error().Error())
		return types.SetNull(types.StringType)
	}
	for _, addon := range *addonsRes.Payload() {
		allDeps = append(allDeps, addon.RealID)
	}

	if len(allDeps) == 0 {
		if stateValue.IsNull() {
			return types.SetNull(types.StringType)
		}
		result, d := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		return result
	}

	result, d := types.SetValueFrom(ctx, types.StringType, allDeps)
	diags.Append(d...)
	return result
}

// SyncDependencies synchronizes all dependencies (both apps and addons) for an application.
// It handles app-to-app dependencies separately from addon dependencies because they use
// different API endpoints.
//
// Expected dependencies formats:
// - "app_xxx" for applications
// - "postgresql_xxx", "addon_xxx" for addons
func SyncDependencies(
	ctx context.Context,
	p provider.Provider,
	applicationID string,
	deps types.Set,
	diags *diag.Diagnostics,
) {
	if deps.IsNull() || deps.IsUnknown() {
		return
	}

	var expectedDeps []string
	diags.Append(deps.ElementsAs(ctx, &expectedDeps, false)...)
	if diags.HasError() {
		return
	}

	// Split dependencies into app IDs and addon IDs
	expectedAppDeps, expectedAddonDeps := splitDependencies(expectedDeps)

	tflog.Debug(ctx, "SYNC DEPENDENCIES", map[string]any{
		"expected": expectedDeps,
		"apps":     expectedAppDeps,
		"addons":   expectedAddonDeps,
	})

	// Sync app dependencies
	syncAppDependencies(ctx, p.Client(), p.Organization(), applicationID, expectedAppDeps, diags)

	// Sync addon dependencies
	syncAddonDependencies(ctx, p.Client(), p.Organization(), applicationID, expectedAddonDeps, diags)
}

// syncAppDependencies syncs app-to-app dependencies using the /dependencies endpoint
func syncAppDependencies(
	ctx context.Context,
	cc *client.Client,
	organization string,
	applicationID string,
	expectedAppDeps []string,
	diags *diag.Diagnostics,
) {
	// Get current linked apps from API
	currentAppsRes := tmp.GetAppDependencies(ctx, cc, organization, applicationID)
	if currentAppsRes.HasError() {
		diags.AddError("failed to get linked apps", currentAppsRes.Error().Error())
		return
	}

	// Extract app IDs from current apps
	currentAppIDs := pkg.Map(*currentAppsRes.Payload(), func(app tmp.AppResponse) string {
		return app.ID
	})

	// Create sets for comparison
	expectedSet := set.New(expectedAppDeps...)
	currentSet := set.New(currentAppIDs...)

	tflog.Debug(ctx, "SYNC APP DEPENDENCIES", map[string]any{
		"expected": expectedAppDeps,
		"current":  currentAppIDs,
	})

	// Remove app dependencies that are no longer expected (current - expected)
	for appID := range currentSet.Difference(expectedSet).Iter() {
		tflog.Info(ctx, "unlinking app dependency", map[string]any{"appID": appID})

		deleteRes := tmp.RemoveAppDependency(ctx, cc, organization, applicationID, appID)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to unlink app "+appID, deleteRes.Error().Error())
		}
	}

	// Add new app dependencies (expected - current)
	for appID := range expectedSet.Difference(currentSet).Iter() {
		tflog.Info(ctx, "linking app dependency", map[string]any{"appID": appID})

		addRes := tmp.AddAppDependency(ctx, cc, organization, applicationID, appID)
		if addRes.HasError() {
			diags.AddError("failed to link app "+appID, addRes.Error().Error())
		}
	}
}

// syncAddonDependencies syncs addon dependencies using the /addons endpoint
func syncAddonDependencies(
	ctx context.Context,
	cc *client.Client,
	organization string,
	applicationID string,
	expectedAddonDeps []string,
	diags *diag.Diagnostics,
) {
	// Convert real IDs to addon IDs (e.g., postgresql_xxx -> addon_xxx)
	var expectedAddonIDs []string
	if len(expectedAddonDeps) > 0 {
		var err error
		expectedAddonIDs, err = tmp.RealIDsToAddonIDs(ctx, cc, organization, expectedAddonDeps...)
		if err != nil {
			diags.AddError("failed to get addon IDs", err.Error())
			return
		}
	}

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
	expectedSet := set.New(expectedAddonIDs...)
	currentSet := set.New(currentAddonIDs...)

	tflog.Debug(ctx, "SYNC ADDON DEPENDENCIES", map[string]any{
		"expected": expectedAddonIDs,
		"current":  currentAddonIDs,
	})

	// Remove addon dependencies that are no longer expected (current - expected)
	for addonID := range currentSet.Difference(expectedSet).Iter() {
		tflog.Info(ctx, "unlinking addon", map[string]any{"addonID": addonID})

		deleteRes := tmp.DeleteAppLinkedAddon(ctx, cc, organization, applicationID, addonID)
		if deleteRes.HasError() && !deleteRes.IsNotFoundError() {
			diags.AddError("failed to unlink addon "+addonID, deleteRes.Error().Error())
		}
	}

	// Add new addon dependencies (expected - current)
	for addonID := range expectedSet.Difference(currentSet).Iter() {
		tflog.Info(ctx, "linking addon", map[string]any{"addonID": addonID})

		addRes := tmp.AddAppLinkedAddons(ctx, cc, organization, applicationID, addonID)
		if addRes.HasError() {
			diags.AddError("failed to link addon "+addonID, addRes.Error().Error())
		}
	}
}
