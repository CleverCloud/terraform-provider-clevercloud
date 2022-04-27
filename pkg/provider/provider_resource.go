package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var resources = map[string]tfsdk.ResourceType{}

func AddResource(name string, rt tfsdk.ResourceType) {
	resources[name] = rt
}

// GetResources - Defines provider resources
func (p *Provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return resources, nil
}
