package static

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceStatic struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceStatic() func() resource.Resource {
	return func() resource.Resource {
		return &ResourceStatic{}
	}
}

func (r *ResourceStatic) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_static"
}
