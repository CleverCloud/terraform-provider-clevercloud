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
)

type Redis struct {
	addon.CommonAttributes
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
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			"host":  schema.StringAttribute{Computed: true, MarkdownDescription: "Database host, used to connect to"},
			"port":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Database port"},
			"token": schema.StringAttribute{Computed: true, MarkdownDescription: "Token to authenticate", Sensitive: true},
		}),
	}
}

// The API-to-state mappers for clevercloud_redis follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name, plan, region
// and creation_date for this resource.
func (rd *Redis) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if rd == nil || addon == nil {
		return
	}

	rd.Name = pkg.FromStr(addon.Name)
	rd.Plan = pkg.FromStr(addon.Plan.Slug)
	rd.Region = pkg.FromStr(addon.Region)
	rd.CreationDate = pkg.FromI(addon.CreationDate)
}

// FromEnv maps the connection details, which Redis only exposes as environment
// variables.
//
// Non-destructive, unlike the application runtimes' FromEnv: an add-on has no
// catch-all `environment` attribute for a residue, so there is nothing to pop.
func (rd *Redis) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) {
	if rd == nil || env == nil {
		return
	}

	vars := env.Map()
	rd.Host = pkg.FromStr(vars["REDIS_HOST"])
	rd.Token = pkg.FromStr(vars["REDIS_PASSWORD"])

	// A port that does not parse is reported as absent rather than failing the
	// whole read: it is a Computed attribute nobody configured, so a hard error
	// would leave the practitioner with a resource they cannot refresh.
	if port, err := strconv.ParseInt(vars["REDIS_PORT"], 10, 64); err == nil {
		rd.Port = pkg.FromI(port)
	} else {
		rd.Port = types.Int64Null()
	}
}
