package networkgroup

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceNG struct {
	cc  *client.Client
	org string

	gitAuth *http.BasicAuth
}

func NewResourceNetworkgroup() resource.Resource {
	return &ResourceNG{}
}

func (r *ResourceNG) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_networkgroup"
}
