package v

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceV struct {
	helper.Configurer
}

func NewResourceV() resource.Resource {
	return &ResourceV{}
}

func (r *ResourceV) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_v"
}

func (r *ResourceV) GetVariantSlug() string {
	return "v"
}
