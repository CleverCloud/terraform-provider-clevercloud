package networkgroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceNG struct {
	helper.Configurer
}

func NewResourceNetworkgroup() resource.Resource {
	return &ResourceNG{}
}

func (r *ResourceNG) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_networkgroup"
}
