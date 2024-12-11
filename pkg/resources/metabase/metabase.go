package metabase

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceMetabase struct {
	cc  *client.Client
	org string
}

func NewResourceMetabase() resource.Resource {
	return &ResourceMetabase{}
}

func (r *ResourceMetabase) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_metabase"
}
