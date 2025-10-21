package mongodb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon/common"
)

type MongoDB struct {
	common.AddonBase
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Database types.String `tfsdk:"database"`
	Uri      types.String `tfsdk:"uri"`
}

//go:embed doc.md
var resourceMongoDBDoc string

func (r ResourceMongoDB) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMongoDBDoc,
		Attributes: common.WithAddonCommons(map[string]schema.Attribute{
			// customer provided
			"host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password": schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true},
			"database": schema.StringAttribute{Computed: true, MarkdownDescription: "Database name"},
			"uri":      schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true},
		}),
	}
}
