package play2

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourcePlay2 struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourcePlay2() func() resource.Resource {
	return func() resource.Resource {
		return &ResourcePlay2{}
	}
}

func (r *ResourcePlay2) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_play2"
}
