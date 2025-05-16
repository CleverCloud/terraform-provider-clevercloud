package fsbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceFSBucket struct {
	cc  *client.Client
	org string
}

func NewResourceFSBucket() resource.Resource {
	return &ResourceFSBucket{}
}

func (r *ResourceFSBucket) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_fsbucket"
}
