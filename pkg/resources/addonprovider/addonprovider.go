package addonprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceAddonProvider struct {
	helper.Configurer
}

func NewResourceAddonProvider() resource.Resource {
	return &ResourceAddonProvider{}
}

func (r *ResourceAddonProvider) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_addon_provider"
}
