package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var datasources = map[string]tfsdk.DataSourceType{}

// GetDataSources - Defines provider data sources
func (p *Provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return datasources, nil
}
