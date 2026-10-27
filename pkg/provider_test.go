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
		{
			name: "dedicated plan",
			plan: &models.AddonPlanView{
				ID: "xs_sml",
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("dedicated")},
				},
			},
			expected: true,
		},
		{
			name: "shared plan",
			plan: &models.AddonPlanView{
				ID: "dev",
				Features: []models.AddonFeatureInstanceView{
					{NameCode: Ptr("is-dedicated"), ComputableValue: Ptr("Shared")},
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
