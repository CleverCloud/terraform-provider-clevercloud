package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var resources = []func() resource.Resource{}

func AddResource(fn func() resource.Resource) {
	resources = append(resources, fn)
}

// GetResources - Defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return resources
}
