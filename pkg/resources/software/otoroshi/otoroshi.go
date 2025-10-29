package otoroshi

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceOtoroshi struct {
	helper.Configurer
}

func NewResourceOtoroshi() resource.Resource {
	return &ResourceOtoroshi{}
}

func (r *ResourceOtoroshi) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_otoroshi"
}
