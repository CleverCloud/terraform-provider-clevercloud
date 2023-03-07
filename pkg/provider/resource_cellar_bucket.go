package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceCellarBucket struct {
	cc  *client.Client
	org string
}

func init() {
	AddResource(NewResourceCellarBucket)
}

func NewResourceCellarBucket() resource.Resource {
	return &ResourceCellarBucket{}
}

func (r *ResourceCellarBucket) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_cellar_bucket"
}
