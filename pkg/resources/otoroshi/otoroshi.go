package otoroshi

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceOtoroshi struct {
	cc  *client.Client
	org string
}

func NewResourceOtoroshi() resource.Resource {
	return &ResourceOtoroshi{}
}

func (r *ResourceOtoroshi) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_otoroshi"
}