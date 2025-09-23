package mysql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type ResourceMySQL struct {
	helper.Configurer
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
