package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

// FromApp maps the application view onto the fields every runtime shares.
//
// It takes the AppResponseProvider interface rather than a *tmp.AppResponse,
// which is the one deviation from the convention's "name and type the API
// payload" rule: GetBuildFlavor() encodes the "no separate build means no build
// flavour" rule, which is state logic rather than API shape, and Create, Read
// and Update have to share it. Do not widen that exception — see CONTRIBUTING.md.
func (r *Runtime) FromApp(ctx context.Context, res AppResponseProvider, diags *diag.Diagnostics) {
	if r == nil || res == nil || res.GetApp() == nil {
		return
	}

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
	r.Branch = pkg.FromStr(app.Branch)

	// The prior value goes in so the mapper can tell a null set from an empty one.
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
