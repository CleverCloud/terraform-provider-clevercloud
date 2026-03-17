package addon

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceAddon) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ad := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
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

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}

	ad.ID = pkg.FromStr(res.Payload().ID)
	ad.CreationDate = pkg.FromI(res.Payload().CreationDate)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), res.Payload().ID)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", envRes.Error().Error())
		return
	}

	envAsMap := pkg.Reduce(*envRes.Payload(), map[string]attr.Value{}, func(acc map[string]attr.Value, v tmp.EnvVar) map[string]attr.Value {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})
	ad.Configurations = types.MapValueMust(types.StringType, envAsMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
}

// Create centralizes the common Create logic for all addon resources.
// Returns the addon_xxx ID (needed for SyncNetworkGroups) and diagnostics.
func Create[T AddonPlan](ctx context.Context, r AddonResource, plan T) (addonID string, diags diag.Diagnostics) {
	common := plan.GetCommonPtr()

	tflog.Debug(ctx, "addon.Create()", map[string]any{"provider": r.GetProviderSlug()})

	// Lookup addon provider and plan
	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		diags.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return "", diags
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, r.GetProviderSlug())
	if prov == nil {
		diags.AddError("addon provider not found", fmt.Sprintf("provider %q does not exist, available: %s", r.GetProviderSlug(), strings.Join(pkg.AddonProvidersAsList(*addonsProviders), ", ")))
		return "", diags
	}

	// Lookup plan: if plan is not specified (single-plan addon), use FirstPlan
	var provPlan *tmp.AddonPlan
	planSlug := common.Plan.ValueString()
	if planSlug == "" || common.Plan.IsNull() || common.Plan.IsUnknown() {
		provPlan = prov.FirstPlan()
		if provPlan == nil {
			diags.AddError("no plan available", "addon provider has no plans")
			return "", diags
		}
	} else {
		provPlan = pkg.LookupProviderPlan(prov, planSlug)
		if provPlan == nil {
			diags.AddError("plan not found", "available plans: "+strings.Join(pkg.ProviderPlansAsList(prov), ", "))
			return "", diags
		}
	}

	// Build creation request
	addonReq := tmp.AddonRequest{
		Name:       common.Name.ValueString(),
		Plan:       provPlan.ID,
		ProviderID: r.GetProviderSlug(),
		Region:     common.Region.ValueString(),
		Options:    plan.GetAddonOptions(),
	}

	// Create addon
	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		diags.AddError("failed to create addon", res.Error().Error())
		return "", diags
	}

	addonPayload := res.Payload()
	common.ID = pkg.FromStr(addonPayload.RealID)
	common.CreationDate = pkg.FromI(addonPayload.CreationDate)
	common.Plan = pkg.FromStr(addonPayload.Plan.Slug)

	// Read provider-specific fields (host, port, password...)
	plan.SetFromResponse(ctx, r.Client(), r.Organization(), addonPayload.ID, &diags)
	if diags.HasError() {
		return addonPayload.ID, diags
	}

	// Set defaults for optional fields
	plan.SetDefaults()

	return addonPayload.ID, diags
}
