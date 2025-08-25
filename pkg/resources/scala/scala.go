package scala

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceScala struct {
	helper.Configurer
}

func NewResourceScala() func() resource.Resource {
	return func() resource.Resource {
		return &ResourceScala{}
	}
}

func (r *ResourceScala) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_scala"
}
