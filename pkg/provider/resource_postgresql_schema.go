package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PostgreSQL struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Plan         types.String `tfsdk:"plan"`
	Region       types.String `tfsdk:"region"`
	CreationDate types.Int64  `tfsdk:"creation_date"`
	Host         types.String `tfsdk:"host"`
	Port         types.Int64  `tfsdk:"port"`
	Database     types.String `tfsdk:"database"`
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
}

const resourcePostgresqlDoc = `
Manage [PostgreSQL](https://www.postgresql.org/) product.

See [product specification](https://www.clever-cloud.com/postgresql-hosting/).

`

func (r ResourcePostgreSQL) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name":   schema.StringAttribute{Required: true, MarkdownDescription: "Name of the service"},
			"plan":   schema.StringAttribute{Required: true, MarkdownDescription: "Database size and spec"},
			"region": schema.StringAttribute{Required: true, MarkdownDescription: "Geographical region where the database will be deployed"},

			// provider
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"creation_date": schema.Int64Attribute{Computed: true, MarkdownDescription: "Date of database creation"},
			"host":          schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":          schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"database":      schema.StringAttribute{Computed: true, MarkdownDescription: "Database name on the PostgreSQL server"},
			"user":          schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password":      schema.StringAttribute{Computed: true, MarkdownDescription: "Login password"},
		},
	}
}
