package mysql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_mysql.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view. It is the only source of name, plan,
// region and creation_date: tmp.MySQL carries none of them.
func (my *MySQL) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *MySQL {
	if my == nil || addon == nil {
		return my
	}

	my.Name = pkg.FromStr(addon.Name)
	my.Plan = pkg.FromStr(addon.Plan.Slug)
	my.Region = pkg.FromStr(addon.Region)
	my.CreationDate = pkg.FromI(addon.CreationDate)

	return my
}

// FromMySQL maps the product view: connection details and the feature toggles.
//
// The feature list is sparse — the API omits a toggle it holds no opinion on —
// and all four toggles are Computed, so an absent one has to resolve to its
// fallback and never to a null. A null is what left an imported add-on with a
// non-empty plan (#404); backup additionally carries a schema default, which
// its fallback has to match.
func (my *MySQL) FromMySQL(ctx context.Context, api *tmp.MySQL, diags *diag.Diagnostics) *MySQL {
	if my == nil || api == nil {
		return my
	}

	my.Host = pkg.FromStr(api.Host)
	my.Port = pkg.FromI(int64(api.Port))
	my.Database = pkg.FromStr(api.Database)
	my.User = pkg.FromStr(api.User)
	my.Password = pkg.FromStr(api.Password)
	my.Version = pkg.FromStr(api.Version)
	my.Uri = pkg.FromStr(api.Uri())
	my.ReadOnlyUsers = tmp.FromMySQLReadOnlyUsers(api.ReadOnlyUsers)

	features := pkg.FeaturesOf(api.Features, func(f tmp.MySQLFeature) (string, bool) {
		return f.Name, f.Enabled
	})
	features.Or(&my.Backup, "do-backup", true) // schema default
	features.Or(&my.Encryption, "encryption", false)
	features.Or(&my.DirectHostOnly, "direct-host-only", false)
	features.Or(&my.SkipLogBin, "skip-log-bin", false)

	return my
}
