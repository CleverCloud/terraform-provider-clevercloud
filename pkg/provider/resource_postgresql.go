package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type resourcePostgresqlType struct{}

const PostgreSQLTypeName = "clevercloud_postgresql"

func init() {
	// Register resource
	AddResource(PostgreSQLTypeName, resourcePostgresqlType{})
}

func (r resourcePostgresqlType) NewResource(ctx context.Context, provider tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p := (provider.(*Provider))
	return ResourcePostgreSQL{
		cc:  p.cc,
		org: p.Organisation,
	}, nil
}
