package cellar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_cellar.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (c *Cellar) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Cellar {
	if c == nil || addon == nil {
		return c
	}

	c.Name = pkg.FromStr(addon.Name)
	c.Region = pkg.FromStr(addon.Region)

	return c
}

// FromEnv maps the S3 credentials, which Cellar only exposes as environment
// variables. Non-destructive — see CONTRIBUTING.md.
func (c *Cellar) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) *Cellar {
	if c == nil || env == nil {
		return c
	}

	credentials := s3.FromEnvVars(env)
	c.Host = pkg.FromStr(credentials.Host)
	c.KeyID = pkg.FromStr(credentials.KeyID)
	c.KeySecret = pkg.FromStr(credentials.KeySecret)

	return c
}
