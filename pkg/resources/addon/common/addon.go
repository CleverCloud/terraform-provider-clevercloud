package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceAddon struct {
	helper.Configurer
}

func NewResourceAddon() resource.Resource {
	return &ResourceAddon{}
}

func (r *ResourceAddon) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_addon"
}
