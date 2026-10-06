package addonprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_addon_provider.
// See CONTRIBUTING.md § "API → state mapping".
//
// This resource is the clearest case of "the API is not the authority on a value
// it does not send": the provider read answers neither the credentials
// (password, sso_salt) nor config_vars nor the four URLs, so every mapper here
// assigns only what its payload actually carries and leaves the rest to the plan
// or the prior state. Do not "complete" them.

// FromProvider maps the provider view: its name and the regions it is released in.
func (ap *AddonProvider) FromProvider(ctx context.Context, api *tmp.AddonProviderInfo, diags *diag.Diagnostics) *AddonProvider {
	if ap == nil || api == nil {
		return ap
	}

	ap.Name = pkg.FromStr(api.Name)
	ap.Regions = pkg.FromSetString(api.Regions, diags)

	return ap
}

// apiFeatureToState converts one API feature view to a state Feature. It is
// shared with SyncFeatures, which maps the features it has just created.
func apiFeatureToState(apiFeature tmp.AddonProviderFeatureView) Feature {
	return Feature{
		Name: pkg.FromStr(apiFeature.Name),
		Type: pkg.FromStr(apiFeature.Type),
	}
}

// FromFeatures maps the feature list.
func (ap *AddonProvider) FromFeatures(ctx context.Context, api []tmp.AddonProviderFeatureView, diags *diag.Diagnostics) *AddonProvider {
	if ap == nil || api == nil {
		return ap
	}

	ap.Features = make([]Feature, 0, len(api))
	for _, feature := range api {
		ap.Features = append(ap.Features, apiFeatureToState(feature))
	}

	return ap
}

// FromPlans maps the plan list, each plan with the features carrying a value.
func (ap *AddonProvider) FromPlans(ctx context.Context, api []tmp.AddonProviderPlanView, diags *diag.Diagnostics) *AddonProvider {
	if ap == nil || api == nil {
		return ap
	}

	ap.Plans = make([]Plan, 0, len(api))
	for _, plan := range api {
		ap.Plans = append(ap.Plans, apiPlanToState(plan))
	}

	return ap
}
