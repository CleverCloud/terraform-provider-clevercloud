package materiakv

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

//go:embed resource_materiakv.md
var resourceMateriaKVDoc string

func (r ResourceMateriaKV) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMateriaKVDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name":   schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"region": schema.StringAttribute{MarkdownDescription: "Geographical region where the database will be deployed", Default: stringdefault.StaticString("par")},
			// provider
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
			"host":          schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":          schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"token":         schema.StringAttribute{Computed: true, MarkdownDescription: "Token to authenticate"},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceMateriaKV) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
