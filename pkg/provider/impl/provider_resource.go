package impl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/registry"
)

// Resources - Defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return registry.Resources
}
