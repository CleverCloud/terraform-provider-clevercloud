package metabase

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

type Metabase struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`
	Host   types.String `tfsdk:"host"`
}

//go:embed doc.md
var resourceMetabaseDoc string

func (r ResourceMetabase) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMetabaseDoc,
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "Metabase host, used to connect to"},
		},
	}
}

// The API-to-state mappers for clevercloud_metabase follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of region for this
// resource — the product view carries the name too, but not the region.
func (mb *Metabase) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if mb == nil || addon == nil {
		return
	}

	mb.Name = pkg.FromStr(addon.Name)
	mb.Region = pkg.FromStr(addon.Region)
}

// FromMetabase maps the product view: the access URL.
func (mb *Metabase) FromMetabase(ctx context.Context, api *tmp.Metabase, diags *diag.Diagnostics) {
	if mb == nil || api == nil {
		return
	}

	mb.Host = pkg.FromStr(api.AccessURL)
}
