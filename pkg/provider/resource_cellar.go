package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceCellar struct {
	cc  *client.Client
	org string
}

func init() {
	AddResource(NewResourceCellar)
}

func NewResourceCellar() resource.Resource {
	return &ResourceCellar{}
}

func (r *ResourceCellar) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_cellar"
}
