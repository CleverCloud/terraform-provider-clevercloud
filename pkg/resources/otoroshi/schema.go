package otoroshi

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Otoroshi struct {
	attributes.Addon
	Version        types.String `tfsdk:"version"`
	Configurations types.Map    `tfsdk:"configurations"`
}

//go:embed doc.md
var resourceOtoroshiDoc string

func (r ResourceOtoroshi) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceOtoroshiDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Otoroshi version to deploy",
			},
			"configurations": schema.MapAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Configuration variables exposed by Otoroshi addon",
				ElementType:         types.StringType,
			},
		}),
	}
}

func (r ResourceOtoroshi) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}