package matomo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_matomo.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of region for this
// resource — the product view carries the name too, but not the region.
func (m *Matomo) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Matomo {
	if m == nil || addon == nil {
		return m
	}

	m.Name = pkg.FromStr(addon.Name)
	m.Region = pkg.FromStr(addon.Region)

	return m
}

// FromMatomo maps the product view: the access URL and the version.
func (m *Matomo) FromMatomo(ctx context.Context, api *tmp.Matomo, diags *diag.Diagnostics) *Matomo {
	if m == nil || api == nil {
		return m
	}

	m.Host = pkg.FromStr(api.AccessURL)
	m.Version = pkg.FromStr(api.Version)

	return m
}
