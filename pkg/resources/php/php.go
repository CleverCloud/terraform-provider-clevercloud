package php

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourcePHP struct {
	cc  *client.Client
	org string
}

func NewResourcePHP() resource.Resource {
	return &ResourcePHP{}
}

func (r *ResourcePHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_php"
}
