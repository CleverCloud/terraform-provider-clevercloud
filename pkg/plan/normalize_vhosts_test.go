package plan

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "basic hostname",
			input:    "example.com",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with trailing slash",
			input:    "example.com/",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with trailing dot",
			input:    "example.com.",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "uppercase hostname",
			input:    "EXAMPLE.COM",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with http schema",
			input:    "http://example.com",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with https schema",
			input:    "https://example.com",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with path",
			input:    "example.com/path/to/resource",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with query parameters",
			input:    "example.com?param=value",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "hostname with fragment",
			input:    "example.com#section",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "full URL with schema, path, query, fragment",
			input:    "https://EXAMPLE.COM/path?param=value#section",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
			hasError: false,
		},
		{
			name:     "hostname with spaces",
			input:    "exam ple.com",
			expected: "",
			hasError: true,
		},
		{
			name:     "hostname with internal slash",
			input:    "example/com",
			expected: "example",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeHost(tt.input)
			
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestNormalizeVHostsPlanModifier_PlanModifySet(t *testing.T) {
	tests := []struct {
		name           string
		configValues   []string
		expectedValues []string
		hasError       bool
	}{
		{
			name:           "normalize mixed case and trailing slashes",
			configValues:   []string{"EXAMPLE.COM/", "test.domain.org"},
			expectedValues: []string{"example.com", "test.domain.org"},
			hasError:       false,
		},
		{
			name:           "remove duplicates after normalization",
			configValues:   []string{"example.com", "EXAMPLE.COM/", "example.com."},
			expectedValues: []string{"example.com"},
			hasError:       false,
		},
		{
			name:           "mixed schemas and formats",
			configValues:   []string{"https://example.com/path", "http://test.com", "domain.org."},
			expectedValues: []string{"domain.org", "example.com", "test.com"},
			hasError:       false,
		},
		{
			name:           "empty list",
			configValues:   []string{},
			expectedValues: []string{},
			hasError:       false,
		},
		{
			name:           "invalid hostname",
			configValues:   []string{"example.com", "invalid host name"},
			expectedValues: []string{"example.com"},
			hasError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := NormalizeVHostsPlanModifier{}
			ctx := context.Background()

			configValue := helper.SetFromStrings(ctx, tt.configValues)

			req := planmodifier.SetRequest{
				Path:        path.Root("vhosts"),
				ConfigValue: configValue,
				StateValue:  types.SetNull(types.StringType),
				PlanValue:   configValue,
			}

			resp := &planmodifier.SetResponse{
				PlanValue: req.PlanValue,
			}

			modifier.PlanModifySet(ctx, req, resp)

			if tt.hasError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected error but got none")
				}
				return
			}

			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected error: %v", resp.Diagnostics)
				return
			}

			var actualValues []string
			diags := resp.PlanValue.ElementsAs(ctx, &actualValues, false)
			if diags.HasError() {
				t.Errorf("failed to extract elements: %v", diags)
				return
			}

			if len(actualValues) != len(tt.expectedValues) {
				t.Errorf("expected %d values but got %d: %v", len(tt.expectedValues), len(actualValues), actualValues)
				return
			}

			for i, expected := range tt.expectedValues {
				if actualValues[i] != expected {
					t.Errorf("at index %d: expected %q but got %q", i, expected, actualValues[i])
				}
			}
		})
	}
}

func TestNormalizeVHostsPlanModifier_NullAndUnknownValues(t *testing.T) {
	modifier := NormalizeVHostsPlanModifier{}
	ctx := context.Background()

	t.Run("null config value", func(t *testing.T) {
		req := planmodifier.SetRequest{
			Path:        path.Root("vhosts"),
			ConfigValue: types.SetNull(types.StringType),
			StateValue:  types.SetNull(types.StringType),
			PlanValue:   types.SetNull(types.StringType),
		}

		resp := &planmodifier.SetResponse{
			PlanValue: req.PlanValue,
		}

		modifier.PlanModifySet(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error: %v", resp.Diagnostics)
		}

		if !resp.PlanValue.IsNull() {
			t.Error("expected plan value to remain null")
		}
	})

	t.Run("unknown config value", func(t *testing.T) {
		req := planmodifier.SetRequest{
			Path:        path.Root("vhosts"),
			ConfigValue: types.SetUnknown(types.StringType),
			StateValue:  types.SetNull(types.StringType),
			PlanValue:   types.SetUnknown(types.StringType),
		}

		resp := &planmodifier.SetResponse{
			PlanValue: req.PlanValue,
		}

		modifier.PlanModifySet(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error: %v", resp.Diagnostics)
		}

		if !resp.PlanValue.IsUnknown() {
			t.Error("expected plan value to remain unknown")
		}
	})
}