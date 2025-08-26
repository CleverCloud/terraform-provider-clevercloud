package redis

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceRedis struct {
	helper.Configurer
}

func NewResourceRedis() resource.Resource {
	return &ResourceRedis{}
}

func (r *ResourceRedis) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_redis"
}
