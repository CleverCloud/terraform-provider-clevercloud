package impl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/registry"
)

// DataSources - Defines provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return registry.Datasources
}

// Resources - Defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return registry.Resources
}

func (p *Provider) Actions(_ context.Context) []func() action.Action {
	return registry.Actions
}
