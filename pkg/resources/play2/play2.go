package play2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourcePlay2 struct {
	helper.Configurer
}

func NewResourcePlay2() func() resource.Resource {
	return func() resource.Resource {
		return &ResourcePlay2{}
	}
}

func (r *ResourcePlay2) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_play2"
}
