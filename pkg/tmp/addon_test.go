package tmp

import (
	"testing"
)

func TestAddonPlan_IsDedicated(t *testing.T) {
	tests := []struct {
		name     string
		plan     *AddonPlan
		expected bool
	}{
		{
			name: "dedicated plan",
			plan: &AddonPlan{
				ID:   "xs_sml",
				Name: "XS Small",
				Slug: "xs_sml",
				Features: []AddonPlanFeature{
					{
						NameCode:        "is-dedicated",
						ComputableValue: "dedicated",
					},
				},
			},
			expected: true,
		},
		{
			name: "shared plan",
			plan: &AddonPlan{
				ID:   "dev",
				Name: "DEV",
				Slug: "dev",
				Features: []AddonPlanFeature{
					{
						NameCode:        "is-dedicated",
						ComputableValue: "Shared",
					},
				},
			},
			expected: false,
		},
		{
			// GET /v2/products/addonproviders, mysql-addon m_med / xs_sml / xxs_sml
			name: "MySQL dedicated plan (boolean is-shared encoding)",
			plan: &AddonPlan{
				ID: "test",
				Features: []AddonPlanFeature{
					{NameCode: "is-dedicated", Type: "BOOLEAN_SHARED", Value: "Dedicated", ComputableValue: "false"},
				},
			},
			expected: true,
		},
		{
			// GET /v2/products/addonproviders, mysql-addon dev
			name: "MySQL shared plan (boolean is-shared encoding)",
			plan: &AddonPlan{
				ID: "test",
				Features: []AddonPlanFeature{
					{NameCode: "is-dedicated", Type: "BOOLEAN_SHARED", Value: "Shared", ComputableValue: "true"},
				},
			},
			expected: false,
		},
		{
			// GET /v2/products/addonproviders, postgresql-addon xs_sml
			name: "PostgreSQL dedicated plan",
			plan: &AddonPlan{
				ID: "test",
				Features: []AddonPlanFeature{
					{NameCode: "is-dedicated", Type: "BOOLEAN_SHARED", Value: "Dedicated", ComputableValue: "dedicated"},
				},
			},
			expected: true,
		},
		{
			// GET /v2/products/addonproviders, postgresql-addon dev
			name: "PostgreSQL shared plan",
			plan: &AddonPlan{
				ID: "test",
				Features: []AddonPlanFeature{
					{NameCode: "is-dedicated", Type: "BOOLEAN_SHARED", Value: "Shared", ComputableValue: "shared"},
				},
			},
			expected: false,
		},
		{
			name: "plan without is-dedicated feature",
			plan: &AddonPlan{
				ID:       "test",
				Name:     "Test Plan",
				Slug:     "test",
				Features: []AddonPlanFeature{},
			},
			expected: false,
		},
		{
			name:     "nil plan",
			plan:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.IsDedicated()
			if result != tt.expected {
				t.Errorf("IsDedicated() = %v, want %v", result, tt.expected)
			}
		})
	}
}
