package materiakv

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

type MateriaKV struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	CreationDate types.Int64  `tfsdk:"creation_date"`
	Host         types.String `tfsdk:"host"`
	Port         types.Int64  `tfsdk:"port"`
	Region       types.String `tfsdk:"region"`
	Token        types.String `tfsdk:"token"`
}

//go:embed doc.md
var resourceMateriaKVDoc string

func (r ResourceMateriaKV) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMateriaKVDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			// provider
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
			"host":          schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":          schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"token":         schema.StringAttribute{Computed: true, MarkdownDescription: "Token to authenticate", Sensitive: true},
		},
	}
}

// The API-to-state mappers for clevercloud_materia_kv follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view.
//
// The Materia payload carries neither name nor region — they live on the add-on
// itself — so this is the only source of name, region and creation_date. Without
// it an import leaves name null and the required attribute fails the plan (#447).
func (kv *MateriaKV) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *MateriaKV {
	if kv == nil || addon == nil {
		return kv
	}

	kv.Name = pkg.FromStr(addon.Name)
	kv.Region = pkg.FromStr(addon.Region)
	kv.CreationDate = pkg.FromI(addon.CreationDate)

	return kv
}

// FromMateriaDB maps the product view: the connection details.
func (kv *MateriaKV) FromMateriaDB(ctx context.Context, api *models.MateriaDB, diags *diag.Diagnostics) *MateriaKV {
	if kv == nil || api == nil {
		return kv
	}

	kv.Host = pkg.FromStr(api.Host)
	kv.Port = pkg.FromI(int64(api.Port))
	kv.Token = pkg.FromStr(api.Token)

	return kv
}
