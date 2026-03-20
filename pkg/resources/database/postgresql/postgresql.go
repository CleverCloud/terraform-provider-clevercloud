package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Ensure ResourcePostgreSQL implements required interfaces
var (
	_ resource.Resource               = &ResourcePostgreSQL{}
	_ resource.ResourceWithModifyPlan = &ResourcePostgreSQL{}
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

	// Validate locale option
	if !plan.Locale.IsNull() && !plan.Locale.IsUnknown() && plan.Locale.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeError(
			path.Root("locale"),
			"Locale not supported on shared plans",
			"Locale support is only available on dedicated PostgreSQL plans (not 'dev'). Please either remove the locale option or upgrade to a dedicated plan (e.g., 'xs_sml', 'xxs_sml').",
		)
	}
}
