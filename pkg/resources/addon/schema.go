package addon

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Addon struct {
	attributes.Addon
	ThirdPartyProvider types.String `tfsdk:"third_party_provider"`
	Configurations     types.Map    `tfsdk:"configurations"`
}

//go:embed doc.md
var resourcePostgresqlDoc string

func (r ResourceAddon) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePostgresqlDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{
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

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceAddon) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
