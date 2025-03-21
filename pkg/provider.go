package pkg

import (
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func AddonProvidersAsList(providers []tmp.AddonProvider) []string {
	return Map(providers, func(provider tmp.AddonProvider) string {
		return provider.ID
	})
}

func LookupAddonProvider(providers []tmp.AddonProvider, providerId string) *tmp.AddonProvider {
	return First(providers, func(provider tmp.AddonProvider) bool {
		return provider.ID == providerId
	})
}

func LookupProviderPlan(provider *tmp.AddonProvider, planSlug string) *tmp.AddonPlan {
	if provider == nil {
		return nil
	}

	return First(provider.Plans, func(plan tmp.AddonPlan) bool {
		return strings.EqualFold(plan.Slug, planSlug)
	})
}

func LookupProviderPlanByID(provider *tmp.AddonProvider, planID string) *tmp.AddonPlan {
	if provider == nil {
		return nil
	}

	return First(provider.Plans, func(plan tmp.AddonPlan) bool {
		return strings.EqualFold(plan.ID, planID)
	})
}

func ProviderPlansAsList(provider *tmp.AddonProvider) []string {
	return Map(provider.Plans, func(plan tmp.AddonPlan) string {
		return plan.Slug
	})
}
