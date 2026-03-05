package pkg

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
			ctx := context.Background()
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
