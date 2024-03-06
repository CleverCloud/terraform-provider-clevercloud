package python

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourcePython struct {
	cc  *client.Client
	org string
}

func NewResourcePython() resource.Resource {
	return &ResourcePython{}
}

func (r *ResourcePython) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_python"
}
