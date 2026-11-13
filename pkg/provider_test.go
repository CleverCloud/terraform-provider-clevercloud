package pkg

import (
	"testing"

	"go.clever-cloud.dev/sdk/models"
)

func TestIsDedicatedPlan(t *testing.T) {
	tests := []struct {
		name     string
		plan     *models.AddonPlanView
		expected bool
	}{
		// the two encodings below are the ones GET /v2/products/addonproviders returns
		{
			name: "dedicated plan",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("dedicated"), Value: Ptr("Dedicated")},
				},
			},
			expected: true,
		},
		{
			name: "shared plan",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("shared"), Value: Ptr("Shared")},
				},
			},
			expected: false,
		},
		{
			name: "MySQL dedicated plan, computable_value tells whether it is shared",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("false"), Value: Ptr("Dedicated")},
				},
			},
			expected: true,
		},
		{
			name: "MySQL shared plan, computable_value tells whether it is shared",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("true"), Value: Ptr("Shared")},
				},
			},
			expected: false,
		},
		{
			name: "computable_value wins over the displayed value",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("false"), Value: Ptr("Shared")},
				},
			},
			expected: true,
		},
		{
			name: "computable_value in another case",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("Dedicated")},
				},
			},
			expected: true,
		},
		{
			name: "no computable_value, dedicated by its displayed value",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), Value: Ptr("Dedicated")},
				},
			},
			expected: true,
		},
		{
			name: "no computable_value, shared by its displayed value",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), Value: Ptr("Shared")},
				},
			},
			expected: false,
		},
		{
			name: "unknown computable_value, dedicated by its displayed value",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("yes"), Value: Ptr("Dedicated")},
				},
			},
			expected: true,
		},
		{
			name: "neither computable_value nor displayed value",
			plan: &models.AddonPlanView{
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated")},
				},
			},
			expected: false,
		},
		{
			name:     "plan without is-dedicated feature",
			plan:     &models.AddonPlanView{ID: "test", Features: []models.AddonFeatureInstanceView{}},
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
			if result := IsDedicatedPlan(tt.plan); result != tt.expected {
				t.Errorf("IsDedicatedPlan() = %v, want %v", result, tt.expected)
			}
		})
	}
}
