package networkgroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceNG struct {
	cc  *client.Client
	org string
}

func NewResourceNetworkgroup() resource.Resource {
	return &ResourceNG{}
}

func (r *ResourceNG) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_networkgroup"
}
