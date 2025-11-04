package defaultloadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type DataSourceDefaultLoadBalancer struct {
	helper.DataSourceConfigurer
}

func NewDataSourceDefaultLoadBalancer() datasource.DataSource {
	return &DataSourceDefaultLoadBalancer{}
}

func (d *DataSourceDefaultLoadBalancer) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_loadbalancer"
}
