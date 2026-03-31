package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Ensure ResourcePostgreSQL implements required interfaces
var (
	_ resource.Resource                 = &ResourcePostgreSQL{}
	_ resource.ResourceWithModifyPlan   = &ResourcePostgreSQL{}
	_ resource.ResourceWithUpgradeState = &ResourcePostgreSQL{}
	_ resource.ResourceWithImportState  = &ResourcePostgreSQL{}
)

type ResourcePostgreSQL struct {
	helper.Configurer
	infos             *tmp.PostgresInfos
	dedicatedVersions []string
}

func NewResourcePostgreSQL() resource.Resource {
	return &ResourcePostgreSQL{
		dedicatedVersions: []string{},
	}
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

				// Parse the old state
				type PostgreSQLV0 struct {
					PostgreSQL
					LocaleOld types.Bool `tfsdk:"locale"` // Old bool field
				}

				var oldState PostgreSQLV0
				res.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if res.Diagnostics.HasError() {
					return
				}

				// Create new state with locale as string
				newState := oldState.PostgreSQL

				// Convert old Bool locale to new String locale
				// If old locale was true, try to retrieve actual locale from database
				// If old locale was false or null, default to en_GB
				if !oldState.LocaleOld.IsNull() && !oldState.LocaleOld.IsUnknown() && oldState.LocaleOld.ValueBool() {
					// Locale was enabled, try to get actual value from database
					if !newState.Host.IsNull() && !newState.Port.IsNull() &&
						!newState.Database.IsNull() && !newState.User.IsNull() && !newState.Password.IsNull() {
						locale, err := getLocaleFromDatabase(
							ctx,
							newState.Host.ValueString(),
							newState.Port.ValueInt64(),
							newState.Database.ValueString(),
							newState.User.ValueString(),
							newState.Password.ValueString(),
						)
						if err != nil {
							tflog.Warn(ctx, "Failed to retrieve locale from database during migration, defaulting to en_GB",
								map[string]any{"error": err.Error()})
							newState.Locale = pkg.FromStr("en_GB")
						} else {
							tflog.Debug(ctx, "Retrieved locale from database during migration", map[string]any{"locale": locale})
							newState.Locale = pkg.FromStr(locale)
						}
					} else {
						newState.Locale = pkg.FromStr("en_GB")
					}
				} else {
					// Locale was false or not set, default to en_GB
					newState.Locale = pkg.FromStr("en_GB")
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

	// Get addon providers to check plan features
	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
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

	isDedicated := addonPlan.IsDedicated()
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
