// Package addonprovider implements the Terraform resource for managing Clever Cloud addon providers.
// This resource allows registering services as addon providers in the Clever Cloud marketplace.
package addonprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// apiFeatureToState converts an API feature response to a state Feature
func apiFeatureToState(apiFeature tmp.AddonProviderFeatureView) Feature {
	return Feature{
		Name: pkg.FromStr(apiFeature.Name),
		Type: pkg.FromStr(apiFeature.Type),
	}
}

// apiPlanToState converts an API plan response to a state Plan
func apiPlanToState(apiPlan tmp.AddonProviderPlanView) Plan {
	// Convert API features to state features
	planFeatures := make([]PlanFeature, 0, len(apiPlan.Features))
	for _, apiFeature := range apiPlan.Features {
		// Only include features that have a non-empty value
		if apiFeature.Value != "" {
			planFeatures = append(planFeatures, PlanFeature{
				Name:  pkg.FromStr(apiFeature.Name),
				Value: pkg.FromStr(apiFeature.Value),
			})
		}
	}

	return Plan{
		ID:       pkg.FromStr(apiPlan.ID),
		Name:     pkg.FromStr(apiPlan.Name),
		Slug:     pkg.FromStr(apiPlan.Slug),
		Price:    pkg.FromFloat64(apiPlan.Price),
		Features: planFeatures,
	}
}

// SyncFeatures synchronizes addon provider features by:
// 1. Reading current features from API
// 2. Deleting features that are no longer in expectedFeatures
// 3. Creating new features from expectedFeatures
//
// This function is called during both Create and Update operations.
func SyncFeatures(
	ctx context.Context,
	cc *client.Client,
	organization string,
	providerID string,
	expectedFeatures []Feature,
	diags *diag.Diagnostics,
) []Feature {
	// Step 1: Get current features from API
	currentFeaturesRes := tmp.ListAddonProviderFeatures(ctx, cc, organization, providerID)
	var currentFeatureNames []string
	if currentFeaturesRes.HasError() && !currentFeaturesRes.IsNotFoundError() {
		diags.AddError("failed to list addon provider features", currentFeaturesRes.Error().Error())
		return nil
	} else if !currentFeaturesRes.IsNotFoundError() {
		for _, f := range *currentFeaturesRes.Payload() {
			currentFeatureNames = append(currentFeatureNames, f.Name)
		}
	}

	// Extract expected feature names
	expectedFeatureNames := make([]string, 0, len(expectedFeatures))
	expectedFeatureMap := make(map[string]Feature)
	for _, f := range expectedFeatures {
		name := f.Name.ValueString()
		expectedFeatureNames = append(expectedFeatureNames, name)
		expectedFeatureMap[name] = f
	}

	// Create sets for comparison
	expectedSet := set.New(expectedFeatureNames...)
	currentSet := set.New(currentFeatureNames...)

	tflog.Debug(ctx, "SYNC FEATURES", map[string]any{
		"expected": expectedFeatureNames,
		"current":  currentFeatureNames,
	})

	// Step 2: Delete features that are no longer expected (current - expected)
	for featureName := range currentSet.Difference(expectedSet).Iter() {
		tflog.Debug(ctx, "deleting feature", map[string]any{"name": featureName})

		delRes := tmp.DeleteAddonProviderFeature(ctx, cc, organization, providerID, featureName)
		if delRes.HasError() && !delRes.IsNotFoundError() {
			diags.AddError(
				fmt.Sprintf("failed to delete feature %s", featureName),
				delRes.Error().Error(),
			)
		}
	}

	// Step 3: Create new features (expected - current)
	resultFeatures := make([]Feature, 0, len(expectedFeatures))
	for featureName := range expectedSet.Difference(currentSet).Iter() {
		feature := expectedFeatureMap[featureName]
		tflog.Info(ctx, "creating feature", map[string]any{"name": featureName})

		featureReq := tmp.AddonProviderFeature{
			Name: feature.Name.ValueString(),
			Type: feature.Type.ValueString(),
		}

		featureRes := tmp.CreateAddonProviderFeature(ctx, cc, organization, providerID, featureReq)
		if featureRes.HasError() {
			diags.AddError(
				fmt.Sprintf("failed to create feature %s", featureName),
				featureRes.Error().Error(),
			)
		} else {
			resultFeatures = append(resultFeatures, apiFeatureToState(*featureRes.Payload()))
		}
	}

	// Add existing features that are still expected
	for featureName := range expectedSet.Intersection(currentSet).Iter() {
		resultFeatures = append(resultFeatures, expectedFeatureMap[featureName])
	}

	return resultFeatures
}

// SyncPlans synchronizes addon provider plans by:
// 1. Reading current plans from API
// 2. Deleting plans that are no longer in expectedPlans
// 3. Creating new plans from expectedPlans
// 4. Updating existing plans that have changed
//
// This function is called during both Create and Update operations.
// It requires the current features to build plan feature types.
func SyncPlans(
	ctx context.Context,
	cc *client.Client,
	organization string,
	providerID string,
	currentFeatures []Feature,
	expectedPlans []Plan,
	diags *diag.Diagnostics,
) []Plan {
	// Build feature types map (needed for plan features)
	featureTypes := make(map[string]string)
	for _, feature := range currentFeatures {
		featureTypes[feature.Name.ValueString()] = feature.Type.ValueString()
	}

	// Step 1: Get current plans from API
	currentPlansRes := tmp.ListAddonProviderPlans(ctx, cc, organization, providerID)
	var currentPlanSlugs []string
	currentPlanMap := make(map[string]tmp.AddonProviderPlanView)
	if currentPlansRes.HasError() && !currentPlansRes.IsNotFoundError() {
		diags.AddError("failed to list addon provider plans", currentPlansRes.Error().Error())
		return nil
	} else if !currentPlansRes.IsNotFoundError() {
		for _, p := range *currentPlansRes.Payload() {
			currentPlanSlugs = append(currentPlanSlugs, p.Slug)
			currentPlanMap[p.Slug] = p
		}
	}

	// Extract expected plan slugs and build map
	expectedPlanSlugs := make([]string, 0, len(expectedPlans))
	expectedPlanMap := make(map[string]Plan)
	for _, p := range expectedPlans {
		slug := p.Slug.ValueString()
		expectedPlanSlugs = append(expectedPlanSlugs, slug)
		expectedPlanMap[slug] = p
	}

	// Create sets for comparison
	expectedSet := set.New(expectedPlanSlugs...)
	currentSet := set.New(currentPlanSlugs...)

	tflog.Debug(ctx, "SYNC PLANS", map[string]any{
		"expected": expectedPlanSlugs,
		"current":  currentPlanSlugs,
	})

	// Step 2: Delete plans that are no longer expected (current - expected)
	for planSlug := range currentSet.Difference(expectedSet).Iter() {
		currentPlan := currentPlanMap[planSlug]
		tflog.Info(ctx, "deleting plan", map[string]any{"slug": planSlug, "id": currentPlan.ID})

		delRes := tmp.DeleteAddonProviderPlan(ctx, cc, organization, providerID, currentPlan.ID)
		if delRes.HasError() && !delRes.IsNotFoundError() {
			diags.AddError(
				fmt.Sprintf("failed to delete plan %s", planSlug),
				delRes.Error().Error(),
			)
		}
	}

	// Helper function to build plan request from a Plan
	buildPlanRequest := func(plan Plan) tmp.AddonProviderPlan {
		planFeatures := make([]tmp.AddonProviderPlanFeature, 0, len(plan.Features))
		for _, feature := range plan.Features {
			featureName := feature.Name.ValueString()
			featureType := featureTypes[featureName]

			planFeatures = append(planFeatures, tmp.AddonProviderPlanFeature{
				Name:  featureName,
				Value: feature.Value.ValueString(),
				Type:  &featureType,
			})
		}

		return tmp.AddonProviderPlan{
			Name:     plan.Name.ValueString(),
			Slug:     plan.Slug.ValueString(),
			Price:    plan.Price.ValueFloat64(),
			Features: planFeatures,
		}
	}

	// Step 3: Create new plans (expected - current)
	resultPlans := make([]Plan, 0, len(expectedPlans))
	for planSlug := range expectedSet.Difference(currentSet).Iter() {
		plan := expectedPlanMap[planSlug]
		tflog.Info(ctx, "creating plan", map[string]any{"slug": planSlug})

		planReq := buildPlanRequest(plan)
		planRes := tmp.CreateAddonProviderPlan(ctx, cc, organization, providerID, planReq)
		if planRes.HasError() {
			diags.AddError(
				fmt.Sprintf("failed to create plan %s", planSlug),
				planRes.Error().Error(),
			)
			continue
		}

		resultPlans = append(resultPlans, apiPlanToState(*planRes.Payload()))
	}

	// Step 4: Update existing plans that are still expected (only if they've changed)
	for planSlug := range expectedSet.Intersection(currentSet).Iter() {
		expectedPlan := expectedPlanMap[planSlug]
		currentPlan := currentPlanMap[planSlug]
		currentStatePlan := apiPlanToState(currentPlan)

		if currentStatePlan.Equal(expectedPlan) {
			tflog.Debug(ctx, "plan unchanged, skipping update", map[string]any{"slug": planSlug})
			resultPlans = append(resultPlans, currentStatePlan)
			continue
		}

		tflog.Info(ctx, "updating plan", map[string]any{"slug": planSlug, "id": currentPlan.ID})

		planReq := buildPlanRequest(expectedPlan)
		planRes := tmp.UpdateAddonProviderPlan(ctx, cc, organization, providerID, currentPlan.ID, planReq)
		if planRes.HasError() {
			diags.AddError(
				fmt.Sprintf("failed to update plan %s", planSlug),
				planRes.Error().Error(),
			)
			continue
		}

		resultPlans = append(resultPlans, apiPlanToState(*planRes.Payload()))

	}

	return resultPlans
}

// Create a new addon provider
func (r *ResourceAddonProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := helper.PlanFrom[AddonProvider](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	configVars := pkg.SetToStringSlice(ctx, plan.ConfigVars, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	manifest := tmp.AddonProviderManifest{
		ID:   plan.ProviderID.ValueString(),
		Name: plan.Name.ValueString(),
		API: tmp.AddonProviderManifestAPIConfig{
			ConfigVars: configVars,
			Password:   plan.Password.ValueString(),
			SSOSalt:    plan.SSOSalt.ValueString(),
			Production: tmp.AddonProviderManifestEnvironmentConfig{
				BaseURL: plan.ProductionBaseURL.ValueString(),
				SSOURL:  plan.ProductionSSOURL.ValueString(),
			},
			Test: tmp.AddonProviderManifestEnvironmentConfig{
				BaseURL: plan.TestBaseURL.ValueString(),
				SSOURL:  plan.TestSSOURL.ValueString(),
			},
		},
	}

	res := tmp.CreateAddonProvider(ctx, r.Client(), r.Organization(), manifest)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create addon provider", res.Error().Error())
		return
	}
	apiProvider := res.Payload()

	// Build initial state from API response
	// The API returns basic info (ID, name, regions) but not the actual config (config_vars, sensitive fields, URLs)
	state := AddonProvider{
		ProviderID:        pkg.FromStr(apiProvider.ID),
		Name:              pkg.FromStr(apiProvider.Name),
		ConfigVars:        plan.ConfigVars,                                           // Not returned by API
		Regions:           pkg.FromSetString(apiProvider.Regions, &resp.Diagnostics), // Computed from API
		Password:          plan.Password,                                             // Not returned by API (sensitive)
		SSOSalt:           plan.SSOSalt,                                              // Not returned by API (sensitive)
		ProductionBaseURL: plan.ProductionBaseURL,                                    // Not returned by API
		ProductionSSOURL:  plan.ProductionSSOURL,                                     // Not returned by API
		TestBaseURL:       plan.TestBaseURL,                                          // Not returned by API
		TestSSOURL:        plan.TestSSOURL,                                           // Not returned by API
		Features:          []Feature{},                                               // Start empty
		Plans:             []Plan{},                                                  // Start empty
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Features = SyncFeatures(ctx, r.Client(), r.Organization(), plan.ProviderID.ValueString(), plan.Features, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Plans = SyncPlans(ctx, r.Client(), r.Organization(), plan.ProviderID.ValueString(), state.Features, plan.Plans, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Read resource information
func (r *ResourceAddonProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ap := helper.StateFrom[AddonProvider](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to get current state
	res := tmp.GetAddonProvider(ctx, r.Client(), r.Organization(), ap.ProviderID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to read addon provider", res.Error().Error())
		return
	}

	// Read features from API
	featuresRes := tmp.ListAddonProviderFeatures(ctx, r.Client(), r.Organization(), ap.ProviderID.ValueString())
	if featuresRes.HasError() && !featuresRes.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to read addon provider features", featuresRes.Error().Error())
		return
	} else if featuresRes.IsNotFoundError() {
		ap.Features = []Feature{}
	} else {
		apiFeatures := *featuresRes.Payload()
		ap.Features = make([]Feature, 0, len(apiFeatures))
		for _, apiFeature := range apiFeatures {
			ap.Features = append(ap.Features, apiFeatureToState(apiFeature))
		}
	}

	// Read plans from API
	plansRes := tmp.ListAddonProviderPlans(ctx, r.Client(), r.Organization(), ap.ProviderID.ValueString())
	if plansRes.HasError() && !plansRes.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to read addon provider plans", plansRes.Error().Error())
		return
	} else if plansRes.IsNotFoundError() {
		ap.Plans = []Plan{}
	} else {
		apiPlans := *plansRes.Payload()
		ap.Plans = make([]Plan, 0, len(apiPlans))
		for _, apiPlan := range apiPlans {
			ap.Plans = append(ap.Plans, apiPlanToState(apiPlan))
		}
	}

	// Keep the current state - the API returns provider info but we can't reconstruct
	// sensitive fields (password, sso_salt) from the response
	resp.Diagnostics.Append(resp.State.Set(ctx, ap)...)
}

// Update resource
func (r *ResourceAddonProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[AddonProvider](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[AddonProvider](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract config_vars set
	configVars := pkg.SetToStringSlice(ctx, plan.ConfigVars, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the JSON manifest with updated values
	manifest := tmp.AddonProviderManifest{
		ID:   plan.ProviderID.ValueString(),
		Name: plan.Name.ValueString(),
		API: tmp.AddonProviderManifestAPIConfig{
			ConfigVars: configVars,
			Password:   plan.Password.ValueString(),
			SSOSalt:    plan.SSOSalt.ValueString(),
			Production: tmp.AddonProviderManifestEnvironmentConfig{
				BaseURL: plan.ProductionBaseURL.ValueString(),
				SSOURL:  plan.ProductionSSOURL.ValueString(),
			},
			Test: tmp.AddonProviderManifestEnvironmentConfig{
				BaseURL: plan.TestBaseURL.ValueString(),
				SSOURL:  plan.TestSSOURL.ValueString(),
			},
		},
	}

	// Update the addon provider via API
	res := tmp.UpdateAddonProvider(ctx, r.Client(), r.Organization(), plan.ProviderID.ValueString(), manifest)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to update addon provider", res.Error().Error())
		return
	}

	tflog.Info(ctx, "Addon provider updated successfully")

	// Update base fields in state
	apiProvider := res.Payload()
	state.ProviderID = plan.ProviderID
	state.Name = plan.Name
	state.ConfigVars = plan.ConfigVars                                        // Not returned by API
	state.Regions = pkg.FromSetString(apiProvider.Regions, &resp.Diagnostics) // Computed from API
	state.Password = plan.Password
	state.SSOSalt = plan.SSOSalt
	state.ProductionBaseURL = plan.ProductionBaseURL
	state.ProductionSSOURL = plan.ProductionSSOURL
	state.TestBaseURL = plan.TestBaseURL
	state.TestSSOURL = plan.TestSSOURL

	state.Features = SyncFeatures(ctx, r.Client(), r.Organization(), plan.ProviderID.ValueString(), plan.Features, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Plans = SyncPlans(ctx, r.Client(), r.Organization(), plan.ProviderID.ValueString(), state.Features, plan.Plans, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceAddonProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ap := helper.StateFrom[AddonProvider](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := tmp.DeleteAddonProvider(ctx, r.Client(), r.Organization(), ap.ProviderID.ValueString())
	if res.HasError() && !res.IsNotFoundError() {
		resp.Diagnostics.AddError("failed to delete addon provider", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
