package metabase

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_metabase.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of region for this
// resource — the product view carries the name too, but not the region.
func (mb *Metabase) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Metabase {
	if mb == nil || addon == nil {
		return mb
	}

	mb.Name = pkg.FromStr(addon.Name)
	mb.Region = pkg.FromStr(addon.Region)

	return mb
}

// FromMetabase maps the product view: the access URL.
func (mb *Metabase) FromMetabase(ctx context.Context, api *tmp.Metabase, diags *diag.Diagnostics) *Metabase {
	if mb == nil || api == nil {
		return mb
	}

	mb.Host = pkg.FromStr(api.AccessURL)

	return mb
}
