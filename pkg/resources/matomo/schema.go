package matomo

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Matomo struct {
	attributes.Addon
	Host types.String `tfsdk:"host"`
	// ID           types.String `tfsdk:"id"`
	// Name         types.String `tfsdk:"name"`
	// CreationDate types.Int64  `tfsdk:"creation_date"`
	// Region       types.String `tfsdk:"region"`
}

//go:embed doc.md
var resourceMatomoDoc string

func (r ResourceMatomo) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMatomoDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{
			"host": schema.StringAttribute{Computed: true, MarkdownDescription: "URL to access Matomo"},
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

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceMatomo) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
