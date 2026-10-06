package fsbucket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_fsbucket.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (fs *FSBucket) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *FSBucket {
	if fs == nil || addon == nil {
		return fs
	}

	fs.Name = pkg.FromStr(addon.Name)
	fs.Region = pkg.FromStr(addon.Region)

	return fs
}

// FromEnv maps the FTP credentials, which the bucket only exposes as
// environment variables. Non-destructive — see CONTRIBUTING.md.
func (fs *FSBucket) FromEnv(ctx context.Context, env tmp.EnvVars, diags *diag.Diagnostics) *FSBucket {
	if fs == nil || env == nil {
		return fs
	}

	vars := env.Map()
	fs.Host = pkg.FromStr(vars["BUCKET_HOST"])
	fs.FTPUsername = pkg.FromStr(vars["BUCKET_FTP_USERNAME"])
	fs.FTPPassword = pkg.FromStr(vars["BUCKET_FTP_PASSWORD"])

	return fs
}
