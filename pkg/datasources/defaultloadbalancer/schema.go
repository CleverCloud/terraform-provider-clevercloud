package defaultloadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DefaultLoadBalancer struct {
	ApplicationID types.String `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	CNAME         types.String `tfsdk:"cname"`
	Servers       types.List   `tfsdk:"servers"`
}

func (d *DataSourceDefaultLoadBalancer) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about the default load balancer for a Clever Cloud application",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "The application ID for which to fetch the load balancer",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the load balancer (usually matches the region)",
			},
			"cname": schema.StringAttribute{
				Computed:    true,
				Description: "The CNAME record for the load balancer",
			},
			"servers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of A records (IP addresses) for the load balancer",
			},
		},
	}
}
