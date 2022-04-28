package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func (r resourcePostgresqlType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: map[string]tfsdk.Attribute{
			// customer provided
			"name":   {Type: types.StringType, Required: true, MarkdownDescription: "Name of the service"},
			"plan":   {Type: types.StringType, Required: true, MarkdownDescription: "Database size and spec"},
			"region": {Type: types.StringType, Required: true, MarkdownDescription: "Geographical region where the database will be deployed"},

			// provider
			"id":            {Type: types.StringType, Computed: true, MarkdownDescription: "Generated unique identifier"},
			"creation_date": {Type: types.Int64Type, Computed: true, MarkdownDescription: "Date of database creation"},
			"host":          {Type: types.StringType, Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":          {Type: types.Int64Type, Computed: true, MarkdownDescription: "Database port"},
			"database":      {Type: types.StringType, Computed: true, MarkdownDescription: "Database name on the PostgreSQL server"},
			"user":          {Type: types.StringType, Computed: true, MarkdownDescription: "Login username"},
			"password":      {Type: types.StringType, Computed: true, Sensitive: true, MarkdownDescription: "Login password"},
		},
	}, nil
}
