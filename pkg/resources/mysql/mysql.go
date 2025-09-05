package mysql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ResourceMySQL struct {
	cc                *client.Client
	org               string
	infos             *tmp.MysqlInfos
	dedicatedVersions []string
}

func NewResourceMySQL() resource.Resource {
	return &ResourceMySQL{
		dedicatedVersions: []string{},
	}
}

func (r *ResourceMySQL) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_mysql"
}
