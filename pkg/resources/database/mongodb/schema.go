package mongodb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
)

type MongoDB struct {
	addon.CommonAttributes
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	User           types.String `tfsdk:"user"`
	Password       types.String `tfsdk:"password"`
	Database       types.String `tfsdk:"database"`
	Uri            types.String `tfsdk:"uri"`
	Encryption     types.Bool   `tfsdk:"encryption"`
	DirectHostOnly types.Bool   `tfsdk:"direct_host_only"`
}

//go:embed doc.md
var resourceMongoDBDoc string

func (r ResourceMongoDB) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceMongoDBDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// customer provided
			"host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password": schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true},
			"database": schema.StringAttribute{Computed: true, MarkdownDescription: "Database name"},
			"uri": schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true},
			"encryption": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Encrypt the hard drive at rest",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"direct_host_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Connect directly to the database host, bypassing the reverse proxy. Lower latency but no automatic failover on migration.",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
		}),
	}
}
