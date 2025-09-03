package bucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceCellarBucket struct {
	helper.Configurer
}

func NewResourceCellarBucket() resource.Resource {
	return &ResourceCellarBucket{}
}

func (r *ResourceCellarBucket) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_cellar_bucket"
}
