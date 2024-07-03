package static

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceStatic struct {
	cc  *client.Client
	org string
}

func NewResourceStatic() func() resource.Resource {
	return func() resource.Resource {
		return &ResourceStatic{}
	}
}

func (r *ResourceStatic) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_static"
}
