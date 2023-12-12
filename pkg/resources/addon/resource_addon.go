package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceAddon struct {
	cc  *client.Client
	org string
}

func NewResourceAddon() resource.Resource {
	return &ResourceAddon{}
}

func (r *ResourceAddon) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_addon"
}
