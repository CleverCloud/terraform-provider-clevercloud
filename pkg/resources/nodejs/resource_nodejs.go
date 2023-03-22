package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.dev/client"
)

type ResourceNodeJS struct {
	cc  *client.Client
	org string
}

func init() {
	impl.AddResource(NewResourceNodeJS)
}

func NewResourceNodeJS() resource.Resource {
	return &ResourceNodeJS{}
}

func (r *ResourceNodeJS) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_nodejs"
}
