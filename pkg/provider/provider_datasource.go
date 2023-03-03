package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var datasources = []func() datasource.DataSource{}

// GetDataSources - Defines provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return datasources
}
