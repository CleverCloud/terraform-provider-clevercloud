package pulsar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourcePulsar struct {
	cc  *client.Client
	org string
}

func NewResourcePulsar() resource.Resource {
	return &ResourcePulsar{}
}

func (r *ResourcePulsar) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_pulsar"
}
