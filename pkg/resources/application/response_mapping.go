package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

// SetFromCreateResponse maps API response fields to Runtime fields after Create operation
func (r *Runtime) SetFromCreateResponse(res *CreateRes, ctx context.Context, diags *diag.Diagnostics) {
	r.ID = pkg.FromStr(res.Application.ID)
	r.Name = pkg.FromStr(res.Application.Name)
	r.Description = pkg.FromStr(res.Application.Description)
	r.MinInstanceCount = pkg.FromI(int64(res.Application.Instance.MinInstances))
	r.MaxInstanceCount = pkg.FromI(int64(res.Application.Instance.MaxInstances))
	r.SmallestFlavor = pkg.FromStr(res.Application.Instance.MinFlavor.Name)
	r.BiggestFlavor = pkg.FromStr(res.Application.Instance.MaxFlavor.Name)
	r.BuildFlavor = res.GetBuildFlavor()
	r.Region = pkg.FromStr(res.Application.Zone)
	r.StickySessions = pkg.FromBool(res.Application.StickySessions)
	r.RedirectHTTPS = pkg.FromBool(ToForceHTTPS(res.Application.ForceHTTPS))
	r.DeployURL = pkg.FromStr(res.Application.DeployURL)

	r.VHosts = helper.VHostsFromAPIHosts(ctx, res.Application.Vhosts.AsString(), r.VHosts, diags)
}

// SetFromReadResponse maps Read API response fields to Runtime fields after Read operation
func (r *Runtime) SetFromReadResponse(res *ReadAppRes, ctx context.Context, diags *diag.Diagnostics) {
	r.Name = pkg.FromStr(res.App.Name)
	r.Description = pkg.FromStr(res.App.Description)
	r.MinInstanceCount = pkg.FromI(int64(res.App.Instance.MinInstances))
	r.MaxInstanceCount = pkg.FromI(int64(res.App.Instance.MaxInstances))
	r.SmallestFlavor = pkg.FromStr(res.App.Instance.MinFlavor.Name)
	r.BiggestFlavor = pkg.FromStr(res.App.Instance.MaxFlavor.Name)
	r.BuildFlavor = res.GetBuildFlavor()
	r.Region = pkg.FromStr(res.App.Zone)
	r.StickySessions = pkg.FromBool(res.App.StickySessions)
	r.RedirectHTTPS = pkg.FromBool(ToForceHTTPS(res.App.ForceHTTPS))
	r.DeployURL = pkg.FromStr(res.App.DeployURL)

	r.VHosts = helper.VHostsFromAPIHosts(ctx, res.App.Vhosts.AsString(), r.VHosts, diags)
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
