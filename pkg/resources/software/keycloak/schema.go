package keycloak

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Keycloak struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Region        types.String `tfsdk:"region"`
	Host          types.String `tfsdk:"host"`
	AdminUsername types.String `tfsdk:"admin_username"`
	AdminPassword types.String `tfsdk:"admin_password"`
}

//go:embed doc.md
var resourceKeycloakDoc string

func (r ResourceKeycloak) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKeycloakDoc,
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			"host":           schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Keycloak"},
			"admin_username": schema.StringAttribute{Computed: true, MarkdownDescription: "Initial admin username for Keycloak"},
			"admin_password": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Initial admin password for Keycloak"},
			// "name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			// "region": schema.StringAttribute{
			// 	Optional:            true,
			// 	Computed:            true,
			// 	Default:             stringdefault.StaticString("par"),
			// 	MarkdownDescription: "Geographical region where the data will be stored",
			// },
			// "id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			// "creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
		},
	}
}
