package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

// SetFromResponse maps API response fields to Runtime fields
func (r *Runtime) SetFromResponse(res AppResponseProvider, ctx context.Context, diags *diag.Diagnostics) {
	app := res.GetApp()

	r.Name = pkg.FromStr(app.Name)
	r.Description = pkg.FromStr(app.Description)
	r.MinInstanceCount = pkg.FromI(int64(app.Instance.MinInstances))
	r.MaxInstanceCount = pkg.FromI(int64(app.Instance.MaxInstances))
	r.SmallestFlavor = pkg.FromStr(app.Instance.MinFlavor.Name)
	r.BiggestFlavor = pkg.FromStr(app.Instance.MaxFlavor.Name)
	r.BuildFlavor = res.GetBuildFlavor()
	r.Region = pkg.FromStr(app.Zone)
	r.StickySessions = pkg.FromBool(app.StickySessions)
	r.RedirectHTTPS = pkg.FromBool(ToForceHTTPS(app.ForceHTTPS))
	r.DeployURL = pkg.FromStr(app.DeployURL)

	r.VHosts = helper.VHostsFromAPIHosts(ctx, app.Vhosts.AsString(), r.VHosts, diags)
}

// FromForceHTTPS converts a boolean to Clever Cloud's ForceHTTPS enum
// on clever side, it's an enum
func FromForceHTTPS(force bool) string {
	if force {
		return "ENABLED"
	} else {
		return "DISABLED"
	}
}

// ToForceHTTPS converts Clever Cloud's ForceHTTPS enum to a boolean
func ToForceHTTPS(force string) bool {
	return force == "ENABLED"
}
