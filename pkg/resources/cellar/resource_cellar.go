package cellar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceCellar struct {
	cc  *client.Client
	org string
}

func NewResourceCellar() resource.Resource {
	return &ResourceCellar{}
}

func (r *ResourceCellar) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_cellar"
}
