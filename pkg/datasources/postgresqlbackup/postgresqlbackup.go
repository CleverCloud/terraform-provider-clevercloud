package postgresqlbackup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type DataSourcePostgreSQLBackup struct {
	helper.DataSourceConfigurer
}

func NewDataSourcePostgreSQLBackup() datasource.DataSource {
	return &DataSourcePostgreSQLBackup{}
}

func (d *DataSourcePostgreSQLBackup) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_backup"
}
