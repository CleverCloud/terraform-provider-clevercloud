package docker

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceDocker struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceDocker() resource.Resource {
	return &ResourceDocker{}
}

func (r *ResourceDocker) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_docker"
}
