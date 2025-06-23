package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ResourcePostgreSQL struct {
	cc                *client.Client
	org               string
	infos             *tmp.PostgresInfos
	dedicatedVersions []string
}

func NewResourcePostgreSQL() resource.Resource {
	return &ResourcePostgreSQL{
		dedicatedVersions: []string{},
	}
}

func (r *ResourcePostgreSQL) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_postgresql"
}
