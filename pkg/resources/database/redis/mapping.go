package redis

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_redis.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name, plan, region
// and creation_date for this resource.
func (rd *Redis) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Redis {
	if rd == nil || addon == nil {
		return rd
	}

	rd.Name = pkg.FromStr(addon.Name)
	rd.Plan = pkg.FromStr(addon.Plan.Slug)
	rd.Region = pkg.FromStr(addon.Region)
	rd.CreationDate = pkg.FromI(addon.CreationDate)

	return rd
}

// FromEnv maps the connection details, which Redis only exposes as environment
// variables.
//
// Non-destructive, unlike the application runtimes' FromEnv: an add-on has no
// catch-all `environment` attribute for a residue, so there is nothing to pop.
func (rd *Redis) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) *Redis {
	if rd == nil || env == nil {
		return rd
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

	return rd
}
