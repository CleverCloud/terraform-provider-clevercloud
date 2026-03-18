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
	"go.clever-cloud.dev/client"
)

// CreateReq represents the request structure for creating an addon
type CreateReq struct {
	Client       *client.Client
	Organization string
	Addon        tmp.AddonRequest
}

// CreateRes represents the response from creating or updating an addon
type CreateRes struct {
	Addon   tmp.AddonResponse
	AddonID string // addon_xxx format (needed for SyncNetworkGroups)
}

func (r *CreateRes) GetAddon() *tmp.AddonResponse {
	return &r.Addon
}

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

// CreateAddon handles the low-level API calls for creating an addon
func CreateAddon(ctx context.Context, req CreateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	res := tmp.CreateAddon(ctx, req.Client, req.Organization, req.Addon)
	if res.HasError() {
		diags.AddError("failed to create addon", res.Error().Error())
		return nil, diags
	}

	addonPayload := res.Payload()

	return &CreateRes{
		Addon:   *addonPayload,
		AddonID: addonPayload.ID,
	}, diags
}

// Create centralizes the common Create logic for all addon resources.
// Returns the addon_xxx ID (needed for SyncNetworkGroups) and diagnostics.
func Create[T AddonPlan](ctx context.Context, r AddonResource, plan T) (addonID string, diags diag.Diagnostics) {
	common := plan.GetCommonPtr()

	tflog.Debug(ctx, "addon.Create()", map[string]any{"provider": r.GetSlug()})

	// Lookup addon provider and plan
	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		diags.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return "", diags
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, r.GetSlug())
	if prov == nil {
		diags.AddError("addon provider not found", fmt.Sprintf("provider %q does not exist, available: %s", r.GetSlug(), strings.Join(pkg.AddonProvidersAsList(*addonsProviders), ", ")))
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

	// Build CreateReq
	createReq := CreateReq{
		Client:       r.Client(),
		Organization: r.Organization(),
		Addon: tmp.AddonRequest{
			Name:       common.Name.ValueString(),
			Plan:       provPlan.ID,
			ProviderID: r.GetSlug(),
			Region:     common.Region.ValueString(),
			Options:    plan.GetAddonOptions(),
		},
	}

	// Call common CreateAddon function
	createRes, createDiags := CreateAddon(ctx, createReq)
	diags.Append(createDiags...)
	if createRes == nil {
		return "", diags
	}

	// Map common fields from response
	a := createRes.GetAddon()
	common.ID = pkg.FromStr(a.RealID)
	common.CreationDate = pkg.FromI(a.CreationDate)
	common.Plan = pkg.FromStr(a.Plan.Slug)

	// Read provider-specific fields (host, port, password...)
	plan.SetFromResponse(ctx, r.Client(), r.Organization(), createRes.AddonID, &diags)
	if diags.HasError() {
		return createRes.AddonID, diags
	}

	// Set defaults for optional fields
	plan.SetDefaults()

	return createRes.AddonID, diags
}
