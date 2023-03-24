package impl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"go.clever-cloud.com/terraform-provider/pkg/registry"
)

// DataSources - Defines provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return registry.Datasources
}
