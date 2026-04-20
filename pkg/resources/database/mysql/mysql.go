package mysql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Ensure ResourceMySQL implements required interfaces
var (
	_ resource.Resource               = &ResourceMySQL{}
	_ resource.ResourceWithModifyPlan = &ResourceMySQL{}
)

type ResourceMySQL struct {
	helper.Configurer
	infos             *tmp.MysqlInfos
	dedicatedVersions []string
}

func NewResourceMySQL() resource.Resource {
	return &ResourceMySQL{
		dedicatedVersions: []string{},
	}
}

func (r *ResourceMySQL) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_mysql"
}

// ModifyPlan validates that encryption, backup, skip_log_bin, and direct_host_only options are only used with dedicated plans
func (r *ResourceMySQL) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() { // Skip validation when deleting
		return
	}

	plan := helper.From[MySQL](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Skip validation if provider not configured yet
	if r.Client() == nil {
		return
	}

	// Get addon providers to check plan features
	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		res.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "mysql-addon")
	if prov == nil {
		res.Diagnostics.AddError("MySQL provider not found", "Could not find mysql-addon provider")
		return
	}

	addonPlan := pkg.LookupProviderPlan(prov, plan.Plan.ValueString())
	if addonPlan == nil {
		// Plan validation will be handled elsewhere
		return
	}

	isDedicated := addonPlan.IsDedicated()

	if !plan.Encryption.IsNull() && !plan.Encryption.IsUnknown() && plan.Encryption.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeError(
			path.Root("encryption"),
			"Encryption not supported on shared plans",
			"At-rest encryption is only available on dedicated MySQL plans. Please either remove the encryption option or upgrade to a dedicated plan (e.g., 'xs_sml', 'xxs_sml').",
		)
	}

	if !plan.Backup.IsNull() && !plan.Backup.IsUnknown() && !plan.Backup.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeWarning(
			path.Root("backup"),
			"Backup configuration not supported on shared plans",
			"The backup option is not configurable on the 'dev' plan.",
		)
	}

	if !plan.SkipLogBin.IsNull() && !plan.SkipLogBin.IsUnknown() && plan.SkipLogBin.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeError(
			path.Root("skip_log_bin"),
			"Skip log bin not supported on shared plans",
			"The skip_log_bin option is only available on dedicated MySQL plans. Please either remove the skip_log_bin option or upgrade to a dedicated plan (e.g., 'xs_sml', 'xxs_sml').",
		)
	}

	if !plan.DirectHostOnly.IsNull() && !plan.DirectHostOnly.IsUnknown() && plan.DirectHostOnly.ValueBool() && !isDedicated {
		res.Diagnostics.AddAttributeError(
			path.Root("direct_host_only"),
			"Direct host only not supported on shared plans",
			"The direct_host_only option is only available on dedicated MySQL plans. Please either remove the direct_host_only option or upgrade to a dedicated plan (e.g., 'xs_sml', 'xxs_sml').",
		)
	}
}
