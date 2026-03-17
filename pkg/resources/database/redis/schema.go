package redis

import (
	"context"
	_ "embed"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type Redis struct {
	addon.CommonAttributes
	Host  types.String `tfsdk:"host"`
	Port  types.Int64  `tfsdk:"port"`
	Token types.String `tfsdk:"token"`
}

func (rd *Redis) GetCommonPtr() *addon.CommonAttributes {
	return &rd.CommonAttributes
}

func (rd *Redis) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (rd *Redis) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	envRes := tmp.GetAddonEnv(ctx, cc, org, addonID)
	if envRes.HasError() {
		diags.AddError("failed to get Redis connection infos", envRes.Error().Error())
		return
	}

	env := *envRes.Payload()
	envAsMap := pkg.Reduce(env, map[string]types.String{}, func(acc map[string]types.String, v tmp.EnvVar) map[string]types.String {
		acc[v.Name] = pkg.FromStr(v.Value)
		return acc
	})

	port, err := strconv.ParseInt(envAsMap["REDIS_PORT"].ValueString(), 10, 64)
	if err != nil {
		diags.AddError("invalid port received", "expect REDIS_PORT to be an Integer")
		return
	}
	rd.Host = envAsMap["REDIS_HOST"]
	rd.Port = pkg.FromI(port)
	rd.Token = envAsMap["REDIS_PASSWORD"]
}

func (rd *Redis) SetDefaults() {
	// Redis has no optional boolean fields
}

//go:embed doc.md
var resourceRedisDoc string

func (r ResourceRedis) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceRedisDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			"host":  schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"token": schema.StringAttribute{Computed: true, MarkdownDescription: "Token to authenticate", Sensitive: true},
		}),
	}
}
