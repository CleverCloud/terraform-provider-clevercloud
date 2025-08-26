package fsbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type ResourceFSBucket struct {
	helper.Configurer
}

func NewResourceFSBucket() resource.Resource {
	return &ResourceFSBucket{}
}

func (r *ResourceFSBucket) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_fsbucket"
}
