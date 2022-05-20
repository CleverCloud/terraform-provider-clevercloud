package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type DatasourceOrganisation struct {
	cc           *client.Client
	organisation string
}

func NewDatasourceOrganisation(cc *client.Client, organisation string) *DatasourceOrganisation {
	return &DatasourceOrganisation{
		cc:           cc,
		organisation: organisation,
	}
}

func (o *DatasourceOrganisation) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, res *tfsdk.ReadDataSourceResponse) {
	orgRes := tmp.GetOrganisation(ctx, o.cc, o.organisation)
	if orgRes.HasError() {
		res.Diagnostics.AddError("failed to get organisation", orgRes.Error().Error())
		return
	}
	org := orgRes.Payload()

	var orgState Organisation
	res.Diagnostics.Append(req.Config.Get(ctx, &orgState)...)
	if res.Diagnostics.HasError() {
		return
	}

	orgState.ID = fromStr(org.ID)
	orgState.Name = fromStr(org.Name)

	res.Diagnostics.Append(res.State.Set(ctx, orgState)...)
}
