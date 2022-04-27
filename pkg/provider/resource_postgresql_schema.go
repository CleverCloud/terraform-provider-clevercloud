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

func (r resourcePostgresqlType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			// customer provided
			"name":   {Type: types.StringType, Required: true},
			"plan":   {Type: types.StringType, Required: true},
			"region": {Type: types.StringType, Required: true},

			// provider
			"id":            {Type: types.StringType, Computed: true}, // for acceptance test retro compat
			"creation_date": {Type: types.Int64Type, Computed: true},
			"host":          {Type: types.StringType, Computed: true},
			"port":          {Type: types.Int64Type, Computed: true},
			"database":      {Type: types.StringType, Computed: true},
			"user":          {Type: types.StringType, Computed: true},
			"password":      {Type: types.StringType, Computed: true, Sensitive: true},
		},
	}, nil
}
