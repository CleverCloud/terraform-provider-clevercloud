package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type DatasourceOrganisationType struct{}

const DatasourceOrganisationName = "clevercloud_organisation"

func init() {
	AddDatasource(DatasourceOrganisationName, &DatasourceOrganisationType{})
}

func (d *DatasourceOrganisationType) NewDataSource(_ context.Context, provider tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p := (provider.(*Provider))
	return NewDatasourceOrganisation(p.cc, p.Organisation), nil
}
