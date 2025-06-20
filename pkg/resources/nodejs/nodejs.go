package nodejs

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceNodeJS struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceNodeJS() resource.Resource {
	return &ResourceNodeJS{}
}

func (r *ResourceNodeJS) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_nodejs"
}
