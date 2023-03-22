package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.dev/client"
)

type ResourcePHP struct {
	cc  *client.Client
	org string
}

func init() {
	impl.AddResource(NewResourcePHP)
}

func NewResourcePHP() resource.Resource {
	return &ResourcePHP{}
}

func (r *ResourcePHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_php"
}
