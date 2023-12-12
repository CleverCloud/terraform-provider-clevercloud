package addon

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Addon struct {
	ID                 types.String `tfsdk:"id"`
	ThirdPartyProvider types.String `tfsdk:"third_party_provider"`
	Name               types.String `tfsdk:"name"`
	CreationDate       types.Int64  `tfsdk:"creation_date"`
	Plan               types.String `tfsdk:"plan"`
	Region             types.String `tfsdk:"region"`
	Configurations     types.Map    `tfsdk:"configurations"`
}

//go:embed resource_addon.md
var resourcePostgresqlDoc string

func (r ResourceAddon) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name":                 schema.StringAttribute{Required: true, MarkdownDescription: "Name of the addon"},
			"plan":                 schema.StringAttribute{Required: true, MarkdownDescription: "billing plan"},
			"region":               schema.StringAttribute{Required: true, MarkdownDescription: "Geographical region where the addon will be deployed (when relevant)"},
			"third_party_provider": schema.StringAttribute{Required: true, MarkdownDescription: "Provider ID"},

			// provider
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
			"configurations": schema.MapAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Any configuration exposed by the addon",
				ElementType:         types.StringType,
			},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceAddon) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
