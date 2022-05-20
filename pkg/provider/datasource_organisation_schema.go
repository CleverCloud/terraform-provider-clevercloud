package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Organisation struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *DatasourceOrganisationType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description:         "Organisation details",
		MarkdownDescription: "Organisation details (either user_xxx or org_xxx)",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				MarkdownDescription: "The organisation ID",
				Computed:            true,
			},
			"name": {
				Type:                types.StringType,
				MarkdownDescription: "Organisation name",
				Computed:            true,
			},
		},
	}, nil
}
