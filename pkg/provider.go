package pkg

import (
	"strings"

	"go.clever-cloud.dev/sdk/models"
)

func AddonProvidersAsList(providers []models.AddonProviderInfoFullView) []string {
	return Map(providers, func(provider models.AddonProviderInfoFullView) string {
		return provider.ID
	})
}

func LookupAddonProvider(providers []models.AddonProviderInfoFullView, providerId string) *models.AddonProviderInfoFullView {
	return First(providers, func(provider models.AddonProviderInfoFullView) bool {
		return provider.ID == providerId
	})
}

func LookupProviderPlan(provider *models.AddonProviderInfoFullView, planSlug string) *models.AddonPlanView {
	if provider == nil {
		return nil
	}

	return First(provider.Plans, func(plan models.AddonPlanView) bool {
		return strings.EqualFold(Deref(plan.Slug), planSlug)
	})
}

func ProviderPlansAsList(provider *models.AddonProviderInfoFullView) []string {
	return Map(provider.Plans, func(plan models.AddonPlanView) string {
		return Deref(plan.Slug)
	})
}

// FirstPlan returns the first plan of a provider, nil when it has none.
func FirstPlan(provider *models.AddonProviderInfoFullView) *models.AddonPlanView {
	if provider == nil || len(provider.Plans) == 0 {
		return nil
	}
	return &provider.Plans[0]
}

// IsDedicatedPlan reports whether the plan is dedicated, from its "is-dedicated"
// feature. The catalogue encodes its computable_value two ways: "dedicated" or
// "shared" for most add-ons, and for MySQL a boolean telling whether the plan is
// shared, so "false" on a dedicated plan. Both are read. When computable_value is
// neither, the displayed value ("Dedicated" or "Shared") decides.
func IsDedicatedPlan(plan *models.AddonPlanView) bool {
	if plan == nil {
		return false
	}
	for _, feature := range plan.Features {
		if Deref(feature.NameCode) != "is-dedicated" {
			continue
		}
		switch strings.ToLower(Deref(feature.ComputableValue)) {
		case "dedicated", "false":
			return true
		case "shared", "true":
			return false
		}
		return strings.EqualFold(Deref(feature.Value), "dedicated")
	}
	return false
}

// EnvMap turns an environment listing into a map of its values by name.
func EnvMap(env []models.AddonEnvironmentView) map[string]string {
	m := make(map[string]string, len(env))
	for _, item := range env {
		m[item.Name] = item.Value
	}
	return m
}
