package docker

import (
	"context"

	"go.clever-cloud.com/terraform-provider/pkg/helper"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type ResourceDocker struct {
	helper.Configurer
}

func NewResourceDocker() resource.Resource {
	return &ResourceDocker{}
}

func (r *ResourceDocker) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_docker"
}
