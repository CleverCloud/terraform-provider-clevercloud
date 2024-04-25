package mongodb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceMongoDB struct {
	cc  *client.Client
	org string
}

func NewResourceMongoDB() resource.Resource {
	return &ResourceMongoDB{}
}

func (r *ResourceMongoDB) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_mongodb"
}
