package golang

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceGo struct {
	cc  *client.Client
	org string
}

func NewResourceGo() resource.Resource {
	return &ResourceGo{}
}

func (r *ResourceGo) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_go"
}
