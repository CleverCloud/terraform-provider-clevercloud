package cellar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceCellar struct {
	helper.Configurer
}

func NewResourceCellar() resource.Resource {
	return &ResourceCellar{}
}

func (r *ResourceCellar) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_cellar"
}
