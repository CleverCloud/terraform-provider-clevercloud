package defaultloadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (d *DataSourceDefaultLoadBalancer) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	config := helper.PlanFrom[DefaultLoadBalancer](ctx, tfsdk.Plan(req.Config), &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Set default region to "par" if not specified
	appId := config.ApplicationID.ValueString()

	tflog.Debug(ctx, "Reading default load balancer", map[string]any{"application_id": appId})

	// Get load balancer information from API
	loadbalancerRes := tmp.GetLoadBalancer(ctx, d.Client(), d.Organization(), appId)
	if loadbalancerRes.HasError() {
		res.Diagnostics.AddError("Failed to get load balancer", loadbalancerRes.Error().Error())
		return
	}
	loadbalancers := *loadbalancerRes.Payload()

	if len(loadbalancers) == 0 {
		res.Diagnostics.AddError("No loadbalancer for this app", "no default loadbalancers")
	}
	loadbalancer := loadbalancers[0]

	// Map the API response to the datasource state
	config.Name = types.StringValue(loadbalancer.Name)
	config.CNAME = types.StringValue(loadbalancer.DNS.CNAME)

	// Convert A records to types.List
	aRecordValues := make([]types.String, len(loadbalancer.DNS.A))
	for i, ip := range loadbalancer.DNS.A {
		aRecordValues[i] = types.StringValue(ip)
	}
	aRecordsList, diags := types.ListValueFrom(ctx, types.StringType, aRecordValues)
	res.Diagnostics.Append(diags...)
	if res.Diagnostics.HasError() {
		return
	}
	config.Servers = aRecordsList

	tflog.Debug(ctx, "Successfully read load balancer", map[string]any{
		"id": config.Name.ValueString(),
	})

	res.Diagnostics.Append(res.State.Set(ctx, &config)...)
}
