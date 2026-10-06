package configprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_configprovider.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view.
//
// The config-provider env endpoint carries no name, so this is the only source
// of it. Without it an import left the required attribute null and the plan
// failed (#447).
func (cp *ConfigProvider) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *ConfigProvider {
	if cp == nil || addon == nil {
		return cp
	}

	cp.Name = pkg.FromStr(addon.Name)

	return cp
}

// FromEnv maps the configured variables, which are this resource's whole point.
// Non-destructive — see CONTRIBUTING.md.
func (cp *ConfigProvider) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) *ConfigProvider {
	if cp == nil || env == nil {
		return cp
	}

	environment, d := types.MapValueFrom(ctx, types.StringType, env.Map())
	diags.Append(d...)
	if d.HasError() {
		return cp
	}
	cp.Environment = environment

	return cp
}
