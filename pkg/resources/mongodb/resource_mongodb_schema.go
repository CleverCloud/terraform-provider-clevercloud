package mongodb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MongoDB struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Plan         types.String `tfsdk:"plan"`
	Region       types.String `tfsdk:"region"`
	CreationDate types.Int64  `tfsdk:"creation_date"`
	Host         types.String `tfsdk:"host"`
	Port         types.Int64  `tfsdk:"port"`
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
}

//go:embed resource_mongodb.md
var resourceMongoDBDoc string

func (r ResourceMongoDB) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMongoDBDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name":   schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"plan":   schema.StringAttribute{Required: true, MarkdownDescription: "Database size and spec"},
			"region": schema.StringAttribute{Required: true, MarkdownDescription: "Geographical region where the data will be stored"}, // TODO, Default: stringdefault.StaticString("par")},

			// provider
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
			"host":          schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":          schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"user":          schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password":      schema.StringAttribute{Computed: true, MarkdownDescription: "Login password"},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceMongoDB) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
