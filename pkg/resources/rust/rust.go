package rust

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceRust struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourceRust() resource.Resource {
	return &ResourceRust{}
}

func (r *ResourceRust) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_rust"
}

const CC_RUST_FEATURES = "CC_RUST_FEATURES"
