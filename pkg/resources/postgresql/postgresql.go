package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.dev/client"
)

type ResourcePostgreSQL struct {
	resources.Controller[PostgreSQL]
	cc  *client.Client
	org string
}

func NewResourcePostgreSQL() resource.Resource {
	return &ResourcePostgreSQL{}
}

func (r *ResourcePostgreSQL) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_postgresql"
}
