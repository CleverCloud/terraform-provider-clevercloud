package common

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Addon struct {
	AddonBase
	ThirdPartyProvider types.String `tfsdk:"third_party_provider"`
	Configurations     types.Map    `tfsdk:"configurations"`
}

//go:embed doc.md
var resourcePostgresqlDoc string

func (r ResourceAddon) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: WithAddonCommons(map[string]schema.Attribute{
			"third_party_provider": schema.StringAttribute{Required: true, MarkdownDescription: "Provider ID"},
			// provider
			"configurations": schema.MapAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Any configuration exposed by the add-on",
				ElementType:         types.StringType,
			},
		}),
	}
}
