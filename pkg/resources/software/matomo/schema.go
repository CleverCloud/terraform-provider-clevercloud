package matomo

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type Matomo struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	Host    types.String `tfsdk:"host"`
	Version types.String `tfsdk:"version"`
}

//go:embed doc.md
var resourceMatomoDoc string

func (r ResourceMatomo) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMatomoDoc,
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Matomo"},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			"version": schema.StringAttribute{Computed: true, MarkdownDescription: "Current version of Matomo"},
		},
	}
}

// The API-to-state mappers for clevercloud_matomo follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of region for this
// resource — the product view carries the name too, but not the region.
func (m *Matomo) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if m == nil || addon == nil {
		return
	}

	m.Name = pkg.FromStr(addon.Name)
	m.Region = pkg.FromStr(addon.Region)
}

// FromMatomo maps the product view: the access URL and the version.
func (m *Matomo) FromMatomo(ctx context.Context, api *tmp.Matomo, diags *diag.Diagnostics) {
	if m == nil || api == nil {
		return
	}

	m.Host = pkg.FromStr(api.AccessURL)
	m.Version = pkg.FromStr(api.Version)
}
