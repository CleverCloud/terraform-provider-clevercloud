package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_postgresql.
// See CONTRIBUTING.md § "API → state mapping".

// defaultLocale is the locale the platform uses when none is requested at creation
const defaultLocale = "en_GB"

// FromAddon maps the generic add-on view. It is the only source of name, plan,
// region and creation_date: tmp.PostgreSQL carries none of them.
func (pg *PostgreSQL) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *PostgreSQL {
	if pg == nil || addon == nil {
		return pg
	}

	pg.Name = pkg.FromStr(addon.Name)
	pg.Plan = pkg.FromStr(addon.Plan.Slug)
	pg.Region = pkg.FromStr(addon.Region)
	pg.CreationDate = pkg.FromI(addon.CreationDate)

	return pg
}

// FromPostgreSQL maps the product view: connection details, locale and the
// feature toggles.
//
// The feature list is sparse and all three toggles are Computed, so an absent
// one resolves to its fallback and never to a null — a null is what left an
// imported add-on with a non-empty plan (#404). backup carries a schema
// default, which its fallback has to match.
func (pg *PostgreSQL) FromPostgreSQL(ctx context.Context, api *tmp.PostgreSQL, diags *diag.Diagnostics) *PostgreSQL {
	if pg == nil || api == nil {
		return pg
	}

	pg.Host = pkg.FromStr(api.Host)
	pg.Port = pkg.FromI(int64(api.Port))
	pg.Database = pkg.FromStr(api.Database)
	pg.User = pkg.FromStr(api.User)
	pg.Password = pkg.FromStr(api.Password)
	pg.Version = pkg.FromStr(api.Version)
	pg.Uri = pkg.FromStr(api.Uri())

	// The locale is fixed at creation time (LC_COLLATE cannot change afterwards)
	// and exposed by the API. When the API does not return it, keep the value
	// already in state (config or default) and only fall back on import.
	switch {
	case api.Locale != "":
		pg.Locale = pkg.FromStr(api.Locale)
	case pg.Locale.IsNull() || pg.Locale.IsUnknown():
		tflog.Warn(ctx, "API did not return the PostgreSQL locale, using default", map[string]any{"locale": defaultLocale})
		pg.Locale = pkg.FromStr(defaultLocale)
	}

	features := pkg.FeaturesOf(api.Features, func(f tmp.PostgreSQLFeature) (string, bool) {
		return f.Name, f.Enabled
	})
	features.Or(&pg.Backup, "do-backup", true) // schema default
	features.Or(&pg.Encryption, "encryption", false)
	features.Or(&pg.DirectHostOnly, "direct-host-only", false)

	return pg
}
