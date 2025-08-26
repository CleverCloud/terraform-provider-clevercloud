package nodejs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceNodeJS struct {
	helper.Configurer
}

func NewResourceNodeJS() resource.Resource {
	return &ResourceNodeJS{}
}

func (r *ResourceNodeJS) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_nodejs"
}
