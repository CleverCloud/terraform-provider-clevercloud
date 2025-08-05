package frankenphp

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceFrankenPHP struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceFrankenPHP() resource.Resource {
	return &ResourceFrankenPHP{}
}

func (r *ResourceFrankenPHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_frankenphp"
}
