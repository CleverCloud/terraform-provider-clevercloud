package networkgroup

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Networkgroup struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
	Network     types.String `tfsdk:"network"`
}

//go:embed doc.md
var resourcePostgresqlDoc string

func (r ResourceNG) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier"},
			"name":        schema.StringAttribute{Required: true, MarkdownDescription: "Name of the network group"},
			"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Description of the network group"},
			"tags":        schema.SetAttribute{ElementType: types.StringType, Optional: true, MarkdownDescription: "Tags of the network group"},
			"network":     schema.StringAttribute{Computed: true, MarkdownDescription: "Network CIDR of the network group"},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceNG) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
