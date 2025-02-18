package redis

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceRedis struct {
	cc  *client.Client
	org string
}

func NewResourceRedis() resource.Resource {
	return &ResourceRedis{}
}

func (r *ResourceRedis) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_redis"
}
