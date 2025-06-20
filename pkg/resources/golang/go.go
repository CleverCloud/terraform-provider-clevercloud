package golang

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceGo struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceGo() resource.Resource {
	return &ResourceGo{}
}

func (r *ResourceGo) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_go"
}
