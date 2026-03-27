package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// configVarsPrefixValidator validates that all config_vars start with the provider_id prefix
type configVarsPrefixValidator struct{}

func (v configVarsPrefixValidator) Description(ctx context.Context) string {
	return "validates that all config_vars are prefixed with the provider_id (uppercased, dashes replaced by underscores)"
}

func (v configVarsPrefixValidator) MarkdownDescription(ctx context.Context) string {
	return "validates that all config_vars are prefixed with the provider_id (uppercased, dashes replaced by underscores)"
}

func (v configVarsPrefixValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	// If the value is unknown or null, there's nothing to validate
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the provider_id value from the same resource
	var providerID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("provider_id"), &providerID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If provider_id is unknown or null, skip validation
	if providerID.IsUnknown() || providerID.IsNull() {
		return
	}

	// Convert provider_id to the expected prefix format:
	// - Uppercase
	// - Replace dashes with underscores
	expectedPrefix := strings.ToUpper(strings.ReplaceAll(providerID.ValueString(), "-", "_"))

	// Get the config_vars set
	var configVars []types.String
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &configVars, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate each config var
	for _, configVar := range configVars {
		if configVar.IsUnknown() || configVar.IsNull() {
			continue
		}

		varName := configVar.ValueString()
		if !strings.HasPrefix(varName, expectedPrefix) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid config_vars prefix",
				fmt.Sprintf(
					"Config variable %q must start with %q (provider_id %q uppercased with dashes replaced by underscores). "+
						"Example: %q",
					varName,
					expectedPrefix,
					providerID.ValueString(),
					expectedPrefix+"_DATABASE_URL",
				),
			)
		}
	}
}

// ConfigVarsPrefixValidator returns a validator that ensures config_vars are prefixed with provider_id
func ConfigVarsPrefixValidator() validator.Set {
	return configVarsPrefixValidator{}
}
