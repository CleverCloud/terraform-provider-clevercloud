package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceKubernetes struct {
	cc  *client.Client
	org string
}

func NewResourceKubernetes() resource.Resource {
	return &ResourceKubernetes{}
}

func (r *ResourceKubernetes) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_kubernetes"
}
