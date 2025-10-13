package keycloak

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Keycloak struct {
	attributes.Addon
	Host types.String `tfsdk:"host"`
	// ID           types.String `tfsdk:"id"`
	// Name         types.String `tfsdk:"name"`
	// CreationDate types.Int64  `tfsdk:"creation_date"`
	// Region       types.String `tfsdk:"region"`
}

//go:embed doc.md
var resourceKeycloakDoc string

func (r ResourceKeycloak) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKeycloakDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Keycloak"},
			// "name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			// "region": schema.StringAttribute{
			// 	Optional:            true,
			// 	Computed:            true,
			// 	Default:             stringdefault.StaticString("par"),
			// 	MarkdownDescription: "Geographical region where the data will be stored",
			// },
			// "id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			// "creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
		}),
	}
}
