package redis

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Redis struct {
	attributes.Addon
	Host  types.String `tfsdk:"host"`
	Port  types.Int64  `tfsdk:"port"`
	Token types.String `tfsdk:"token"`
}

//go:embed doc.md
var resourceRedisDoc string

func (r ResourceRedis) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceRedisDoc,
		Attributes: attributes.WithAddonCommons(map[string]schema.Attribute{
			"host":  schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"token": schema.StringAttribute{Computed: true, MarkdownDescription: "Token to authenticate"},
		}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceRedis) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
