package nodejs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceNodeJS struct {
	cc  *client.Client
	org string
}

func NewResourceNodeJS() resource.Resource {
	return &ResourceNodeJS{}
}

func (r *ResourceNodeJS) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_nodejs"
}
