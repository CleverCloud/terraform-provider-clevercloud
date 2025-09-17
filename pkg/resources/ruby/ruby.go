package ruby

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceRuby struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceRuby() resource.Resource {
	return &ResourceRuby{}
}

func (r *ResourceRuby) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_ruby"
}