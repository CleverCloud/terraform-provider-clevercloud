package configprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceConfigProvider struct {
	helper.Configurer
}

func NewResourceConfigProvider() resource.Resource {
	return &ResourceConfigProvider{}
}

func (r *ResourceConfigProvider) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_configprovider"
}
