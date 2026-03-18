package pkg

import (
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
			name:        "valid empty map",
			value:       types.MapValueMust(types.StringType, map[string]attr.Value{}),
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
			ctx := t.Context()

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

func TestLowercaseValidator(t *testing.T) {
	tests := []struct {
		name        string
		value       types.String
		expectError bool
	}{
		{
			name:        "valid lowercase value",
			value:       types.StringValue("3xl_cpu_tit"),
			expectError: false,
		},
		{
			name:        "valid all lowercase with numbers",
			value:       types.StringValue("dev"),
			expectError: false,
		},
		{
			name:        "valid lowercase with dashes",
			value:       types.StringValue("xs-cpu"),
			expectError: false,
		},
		{
			name:        "invalid uppercase value",
			value:       types.StringValue("3XL_CPU_TIT"),
			expectError: true,
		},
		{
			name:        "invalid mixed case value",
			value:       types.StringValue("3Xl_Cpu_Tit"),
			expectError: true,
		},
		{
			name:        "invalid partial uppercase",
			value:       types.StringValue("Dev"),
			expectError: true,
		},
		{
			name:        "valid null value (should not error)",
			value:       types.StringNull(),
			expectError: false,
		},
		{
			name:        "valid unknown value (should not error)",
			value:       types.StringUnknown(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			lowercaseValidator := NewLowercaseValidator()

			req := validator.StringRequest{
				Path:           path.Root("plan"),
				PathExpression: path.MatchRoot("plan"),
				ConfigValue:    tt.value,
			}
			res := &validator.StringResponse{}

			lowercaseValidator.ValidateString(ctx, req, res)

			hasError := res.Diagnostics.HasError()
			if hasError != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, hasError)
				if hasError {
					t.Logf("Diagnostics: %v", res.Diagnostics)
				}
			}

			// Additional check: verify error message mentions the lowercase version
			if tt.expectError && !tt.value.IsNull() && !tt.value.IsUnknown() {
				value := tt.value.ValueString()
				expectedLowercase := strings.ToLower(value)

				found := false
				for _, diag := range res.Diagnostics.Errors() {
					if strings.Contains(diag.Detail(), expectedLowercase) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"expected error message to suggest lowercase value '%s', got diagnostics: %v",
						expectedLowercase,
						res.Diagnostics)
				}
			}
		})
	}
}
