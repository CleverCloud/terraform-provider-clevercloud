package static

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceStatic struct {
	helper.Configurer
}

func NewResourceStatic() func() resource.Resource {
	return func() resource.Resource {
		return &ResourceStatic{}
	}
}

func (r *ResourceStatic) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_static"
}
