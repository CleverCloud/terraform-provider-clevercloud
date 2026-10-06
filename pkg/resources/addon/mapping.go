package addon

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_addon.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view.
func (ad *Addon) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Addon {
	if ad == nil || addon == nil {
		return ad
	}

	ad.Name = pkg.FromStr(addon.Name)

	// Providers do not agree on the case of their plan slugs: jenkins answers S,
	// M, L while postgresql answers dev, xs_sml. LookupProviderPlan matches
	// case-insensitively, so both spellings create the same add-on; keep the one
	// already in state so a plan never drifts on case alone. An import has
	// nothing in state yet and takes the provider's spelling.
	plan := addon.Plan.Slug
	if configured := ad.Plan.ValueString(); strings.EqualFold(configured, plan) {
		plan = configured
	}
	ad.Plan = pkg.FromStr(plan)

	ad.Region = pkg.FromStr(addon.Region)
	ad.ThirdPartyProvider = pkg.FromStr(addon.Provider.ID)
	ad.CreationDate = pkg.FromI(addon.CreationDate)

	return ad
}

// FromEnv maps the add-on's whole environment into the configurations map.
// Non-destructive — see CONTRIBUTING.md.
func (ad *Addon) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) *Addon {
	if ad == nil || env == nil {
		return ad
	}

	values := map[string]attr.Value{}
	for name, value := range env.Map() {
		values[name] = pkg.FromStr(value)
	}
	ad.Configurations = types.MapValueMust(types.StringType, values)

	return ad
}
