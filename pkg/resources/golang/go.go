package golang

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceGo struct {
	helper.Configurer
}

func NewResourceGo() resource.Resource {
	return &ResourceGo{}
}

func (r *ResourceGo) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_go"
}
