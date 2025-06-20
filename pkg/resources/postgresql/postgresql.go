package postgresql

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourcePostgreSQL struct {
	cc      *client.Client
	org     string
	gitAuth *http.BasicAuth
}

func NewResourcePostgreSQL() resource.Resource {
	return &ResourcePostgreSQL{}
}

func (r *ResourcePostgreSQL) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_postgresql"
}
