package pkg

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNoNullMapValuesValidator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.Map
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid map with all non-null values",
			value: types.MapValueMust(types.StringType, map[string]attr.Value{
				"VAR1": types.StringValue("value1"),
				"VAR2": types.StringValue("value2"),
			}),
			expectError: false,
		},
		{
			name: "valid empty map",
			value: types.MapValueMust(types.StringType, map[string]attr.Value{}),
			expectError: false,
		},
		{
			name:        "valid null map (not set)",
			value:       types.MapNull(types.StringType),
			expectError: false,
		},
		{
			name:        "valid unknown map",
			value:       types.MapUnknown(types.StringType),
			expectError: false,
		},
		{
			name: "invalid map with one null value",
			value: types.MapValueMust(types.StringType, map[string]attr.Value{
				"VAR1": types.StringValue("value1"),
				"VAR2": types.StringNull(),
			}),
			expectError: true,
			errorMsg:    "VAR2",
		},
		{
			name: "invalid map with multiple null values",
			value: types.MapValueMust(types.StringType, map[string]attr.Value{
				"VAR1": types.StringNull(),
				"VAR2": types.StringValue("value2"),
				"VAR3": types.StringNull(),
			}),
			expectError: true,
			errorMsg:    "VAR1",
		},
		{
			name: "invalid map with all null values",
			value: types.MapValueMust(types.StringType, map[string]attr.Value{
				"VAR1": types.StringNull(),
				"VAR2": types.StringNull(),
			}),
			expectError: true,
			errorMsg:    "VAR1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			req := validator.MapRequest{
				Path:           path.Root("environment"),
				PathExpression: path.MatchRoot("environment"),
				ConfigValue:    tt.value,
			}
			res := &validator.MapResponse{}

			v := NoNullMapValuesValidator()
			v.ValidateMap(ctx, req, res)

			hasError := res.Diagnostics.HasError()
			if hasError != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, hasError)
				if hasError {
					t.Logf("Diagnostics: %v", res.Diagnostics)
				}
			}

			if tt.expectError && len(tt.errorMsg) > 0 {
				found := false
				for _, diag := range res.Diagnostics.Errors() {
					if strings.Contains(diag.Detail(), tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"expected error message containing '%s', got diagnostics: %v",
						tt.errorMsg,
						res.Diagnostics,
					)
				}
			}
		})
	}
}
