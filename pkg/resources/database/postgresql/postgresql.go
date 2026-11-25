package postgresql

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/addonid"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

// Ensure ResourcePostgreSQL implements required interfaces
var (
	_ resource.Resource                 = &ResourcePostgreSQL{}
	_ resource.ResourceWithModifyPlan   = &ResourcePostgreSQL{}
	_ resource.ResourceWithUpgradeState = &ResourcePostgreSQL{}
	_ resource.ResourceWithImportState  = &ResourcePostgreSQL{}
)

type ResourcePostgreSQL struct {
	helper.AddonConfigurer
	// What a new add-on can be created on, and what an existing one can be
	// moved to: the second offers versions that are closed to new add-ons.
	creationInfos  *models.PostgresqlProvidersConsole
	migrationInfos *models.PostgresqlProvidersConsole
}

func NewResourcePostgreSQL() resource.Resource {
	return &ResourcePostgreSQL{}
}

func (r *ResourcePostgreSQL) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_postgresql"
}

// UpgradeState handles schema migrations from older versions
func (r *ResourcePostgreSQL) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Migration from version 0 (locale as Bool) to version 1 (locale as String)
		0: {
			PriorSchema: &schemaPostgresqlV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading PostgreSQL state from version 0 to version 1")

				// Parse the old state using a flat struct matching schema v0 exactly.
				// Cannot embed PostgreSQL here because it already has Locale types.String `tfsdk:"locale"`,
				// which would conflict with the v0 Bool field of the same tag.
				type PostgreSQLV0 struct {
					addon.CommonAttributes
					Host           types.String `tfsdk:"host"`
					Port           types.Int64  `tfsdk:"port"`
					Database       types.String `tfsdk:"database"`
					User           types.String `tfsdk:"user"`
					Password       types.String `tfsdk:"password"`
					Version        types.String `tfsdk:"version"`
					Uri            types.String `tfsdk:"uri"`
					Backup         types.Bool   `tfsdk:"backup"`
					Encryption     types.Bool   `tfsdk:"encryption"`
					DirectHostOnly types.Bool   `tfsdk:"direct_host_only"`
					Locale         types.Bool   `tfsdk:"locale"` // Bool in v0, String in v1
				}

				var oldState PostgreSQLV0
				res.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if res.Diagnostics.HasError() {
					return
				}

				// Build new state by copying all common fields explicitly
				newState := PostgreSQL{
					CommonAttributes: oldState.CommonAttributes,
					Host:             oldState.Host,
					Port:             oldState.Port,
					Database:         oldState.Database,
					User:             oldState.User,
					Password:         oldState.Password,
					Version:          oldState.Version,
					Uri:              oldState.Uri,
					Backup:           oldState.Backup,
					Encryption:       oldState.Encryption,
					DirectHostOnly:   oldState.DirectHostOnly,
				}

				// Convert old Bool locale to new String locale.
				// The locale is exposed by the API: read it from there, defaulting to en_GB
				// when it cannot be resolved (API not reachable, addon gone, older API)
				newState.Locale = pkg.FromStr(defaultLocale)
				if !oldState.Locale.IsNull() && !oldState.Locale.IsUnknown() && oldState.Locale.ValueBool() {
					if locale := r.localeFromAPI(ctx, newState.ID.ValueString()); locale != "" {
						newState.Locale = pkg.FromStr(locale)
					}
				}

				tflog.Info(ctx, "Successfully upgraded PostgreSQL state", map[string]any{"locale": newState.Locale.ValueString()})
				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}

// ModifyPlan validates that encryption, backup, and locale options are only used with dedicated plans
func (r *ResourcePostgreSQL) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	tflog.Debug(ctx, "ModifyPlan called for PostgreSQL")

	if req.Plan.Raw.IsNull() { // Skip validation when deleting
		return
	}

	plan := helper.From[PostgreSQL](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Skip validation if provider not configured yet
	if r.Client() == nil {
		tflog.Debug(ctx, "Skipping validation - provider not configured")
		return
	}

	r.validateVersion(ctx, req, *plan, res)

	// Get addon providers to check plan features
	addonsProvidersRes := pkg.AddonProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		tflog.Warn(ctx, "Skipping validation - failed to get addon providers", map[string]any{"error": addonsProvidersRes.Error().Error()})
		// Silently skip validation if we can't fetch provider info during plan phase
		// The actual creation will fail with a proper error if there's a real issue
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "postgresql-addon")
	if prov == nil {
		tflog.Debug(ctx, "Skipping validation - PostgreSQL provider not found")
		// Skip validation if provider info is not available
		return
	}

	addonPlan := pkg.LookupProviderPlan(prov, plan.Plan.ValueString())
	if addonPlan == nil {
		tflog.Debug(ctx, "Skipping validation - plan not found", map[string]any{"plan": plan.Plan.ValueString()})
		// Plan validation will be handled elsewhere
		return
	}

	isDedicated := pkg.IsDedicatedPlan(addonPlan)
	tflog.Debug(ctx, "Plan validation", map[string]any{"plan": plan.Plan.ValueString(), "isDedicated": isDedicated})

	// Validate encryption option
	if !plan.Encryption.IsNull() && !plan.Encryption.IsUnknown() && plan.Encryption.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeError(
			path.Root("encryption"),
			"Encryption not supported on shared plans",
			"At-rest encryption is only available on dedicated PostgreSQL plans. Please either remove the encryption option or upgrade to a dedicated plan (e.g., 'xs_sml', 'xxs_sml').",
		)
	}

	// Validate backup option
	if !plan.Backup.IsNull() && !plan.Backup.IsUnknown() && !plan.Backup.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeWarning(
			path.Root("backup"),
			"Backup configuration may not be supported on shared plans",
			"The backup option may not be configurable on the 'dev' plan as it uses a shared cluster.",
		)
	}

	// Validate locale option - only check if user explicitly set it in config
	// The provider can read/compute locale for any plan, but users can only SET locale on dedicated plans
	var configLocale types.String
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("locale"), &configLocale)...)
	if res.Diagnostics.HasError() {
		return
	}

	// A shared-plan add-on can already carry a locale the platform gave it, and
	// Read puts that value in state. `terraform plan -generate-config-out` then
	// writes it into the generated configuration, so refusing it outright makes
	// the add-on impossible to import, or even to plan against. Only refuse a
	// locale the practitioner is introducing or changing.
	if !req.State.Raw.IsNull() {
		var stateLocale types.String
		res.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("locale"), &stateLocale)...)
		if res.Diagnostics.HasError() {
			return
		}

		if stateLocale.Equal(configLocale) {
			return
		}
	}

	// Only validate if user explicitly provided a locale value (not null = user set it)
	// If it's null, the default or computed value will be used, which is fine
	if !configLocale.IsNull() && !configLocale.IsUnknown() && !isDedicated {
		res.Diagnostics.AddAttributeError(
			path.Root("locale"),
			"Locale not supported on shared plans",
			"Locale support is only available on dedicated PostgreSQL plans (not 'dev'). Please either remove the locale option or upgrade to a dedicated plan (e.g., 'xs_sml', 'xxs_sml').",
		)
	}
}

// validateVersion refuses a version the platform does not offer. A new add-on
// is checked against the versions open to creation, a version change against
// the ones an existing add-on can be migrated to. The version an add-on already
// runs is never refused: the platform keeps running versions it no longer offers.
func (r *ResourcePostgreSQL) validateVersion(ctx context.Context, req resource.ModifyPlanRequest, plan PostgreSQL, res *resource.ModifyPlanResponse) {
	var version types.String
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("version"), &version)...)
	if res.Diagnostics.HasError() || version.IsNull() || version.IsUnknown() {
		return
	}

	migrating := !req.State.Raw.IsNull()
	if migrating {
		var current types.String
		res.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("version"), &current)...)
		if res.Diagnostics.HasError() || current.Equal(version) {
			return
		}
	}

	infos := r.providerInfos(ctx, migrating)
	if infos == nil { // not reachable: the API refuses a wrong version at apply
		return
	}

	requestVersion := version.ValueString()
	region := plan.Region.ValueString()

	if plan.Plan.ValueString() == "dev" {
		cluster := pkg.First(infos.Clusters, func(cluster models.PostgresqlClusterConsole) bool {
			return cluster.Zone == region
		})
		if cluster == nil {
			res.Diagnostics.AddAttributeError(
				path.Root("version"),
				"No PostgreSQL dev cluster found for this region",
				fmt.Sprintf("could not determine dev cluster on region %s", region),
			)
			return
		}

		if cluster.Version != requestVersion {
			res.Diagnostics.AddAttributeError(
				path.Root("version"),
				"PostgreSQL version not available on this cluster",
				fmt.Sprintf("Cluster %s is running version %s, not version %s", cluster.Label, cluster.Version, requestVersion),
			)
		}
		return
	}

	// on dedicated plan, any available version is OK
	if _, exists := infos.Dedicated[requestVersion]; !exists {
		summary := "PostgreSQL version not available"
		if migrating {
			summary = "PostgreSQL version not available for a migration"
		}
		res.Diagnostics.AddAttributeError(
			path.Root("version"),
			summary,
			fmt.Sprintf(
				"version %s not available, available versions: %s",
				requestVersion,
				strings.Join(slices.Sorted(maps.Keys(infos.Dedicated)), ", "),
			),
		)
	}
}

// localeFromAPI resolves the locale of an addon from the API, given its real ID
// (postgresql_xxx). Returns an empty string when it cannot be resolved.
func (r *ResourcePostgreSQL) localeFromAPI(ctx context.Context, realID string) string {
	if r.Provider == nil || r.Client() == nil {
		return ""
	}

	addonID, err := addonid.FromRealID(ctx, r.Client(), r.Organization(), realID)
	if err != nil {
		tflog.Warn(ctx, "failed to resolve addon ID while reading locale", map[string]any{"error": err.Error()})
		return ""
	}

	res := tmp.GetPostgreSQL(ctx, r.Client(), addonID)
	if res.HasError() {
		tflog.Warn(ctx, "failed to read locale from the API", map[string]any{"error": res.Error().Error()})
		return ""
	}

	return pkg.Deref(res.Payload().Locale)
}
