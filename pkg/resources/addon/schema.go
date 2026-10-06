package addon

import (
	"context"
	_ "embed"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type Addon struct {
	CommonAttributes
	ThirdPartyProvider types.String `tfsdk:"third_party_provider"`
	Configurations     types.Map    `tfsdk:"configurations"`
}

//go:embed doc.md
var resourcePostgresqlDoc string

func (r ResourceAddon) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: WithAddonCommons(map[string]schema.Attribute{
			"third_party_provider": schema.StringAttribute{Required: true, MarkdownDescription: "Provider ID"},
			// provider
			"configurations": schema.MapAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Any configuration exposed by the add-on",
				ElementType:         types.StringType,
			},
		}),
	}
}

// The API-to-state mappers for clevercloud_addon follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view.
func (ad *Addon) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if ad == nil || addon == nil {
		return
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
}

// FromEnv maps the add-on's whole environment into the configurations map.
// Non-destructive — see CONTRIBUTING.md.
func (ad *Addon) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) {
	if ad == nil || env == nil {
		return
	}

	values := map[string]attr.Value{}
	for name, value := range env.Map() {
		values[name] = pkg.FromStr(value)
	}
	ad.Configurations = types.MapValueMust(types.StringType, values)
}
