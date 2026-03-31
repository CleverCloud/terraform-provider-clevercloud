package postgresql

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
)

// schemaPostgresqlV0 represents the schema for version 0 where locale was a Bool
var schemaPostgresqlV0 = schema.Schema{
	Version: 0,
	Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
		"host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
		"port":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
		"database": schema.StringAttribute{Computed: true, MarkdownDescription: "Database name on the PostgreSQL server"},
		"user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
		"password": schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true},
		"uri":      schema.StringAttribute{Computed: true, MarkdownDescription: "Database connection string", Sensitive: true},
		"version": schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "PostgreSQL version",
		},
		"backup": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			MarkdownDescription: "Enable or disable backups for this PostgreSQL add-on.",
		},
		"encryption": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Encrypt the hard drive at rest",
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		"direct_host_only": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Connect directly to the database host",
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		// OLD SCHEMA: locale was a Bool
		"locale": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Enable locale support for collation and character classification",
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
	}),
}
