package pkg

import "github.com/hashicorp/terraform-plugin-framework/types"

// Features is a presence-aware view of the sparse named-toggle lists the add-on
// APIs return (tmp.MySQLFeature, tmp.PostgreSQLFeature, tmp.MongoDBFeature,
// tmp.ElasticsearchFeature and friends).
//
// Absent is not false. The API omits a toggle it holds no opinion on, so the
// right answer for a missing key is the attribute's fallback — never a null, and
// never a blind false. A null on a Computed attribute is the bug class of #404:
// an import leaves it empty, the schema default materialises on the next plan,
// and the plan is never empty. With RequiresReplace on top of that, the plan is
// a destroy and recreate (#452).
//
// It is the feature-list counterpart of SetBool / SetBoolIf: Or is to a toggle
// what SetBool is to an environment variable, Keep what SetBoolIf is.
type Features map[string]bool

// FeaturesOf collects a sparse toggle list into a Features.
//
// split names the two fields rather than reaching for them by reflection, so a
// renamed field is a compile error:
//
//	pkg.FeaturesOf(api.Features, func(f tmp.MySQLFeature) (string, bool) {
//		return f.Name, f.Enabled
//	})
func FeaturesOf[T any](items []T, split func(T) (string, bool)) Features {
	features := make(Features, len(items))
	for _, item := range items {
		name, enabled := split(item)
		features[name] = enabled
	}

	return features
}

// Or assigns target from the API, resolving an absent key to fallback.
//
// This is what every Computed attribute wants, and the only policy that
// survives a `terraform import`. Where the attribute carries a schema Default,
// fallback must be that same default.
func (f Features) Or(target *types.Bool, key string, fallback bool) {
	if enabled, ok := f[key]; ok {
		*target = types.BoolValue(enabled)
		return
	}

	*target = types.BoolValue(fallback)
}

// Keep leaves target untouched when the key is absent, preserving whatever the
// plan or the prior state held.
//
// Only correct for an Optional attribute with no schema default and no Computed
// flag, where a null still means "the practitioner never configured this". Same
// trade-off as SetBoolIf against SetBool.
func (f Features) Keep(target *types.Bool, key string) {
	if enabled, ok := f[key]; ok {
		*target = types.BoolValue(enabled)
	}
}

// Null writes a null when the key is absent, for an attribute whose emptiness
// is meaningful and which carries no default to fall back on.
func (f Features) Null(target *types.Bool, key string) {
	if enabled, ok := f[key]; ok {
		*target = types.BoolValue(enabled)
		return
	}

	*target = types.BoolNull()
}

// Enabled is the raw presence-aware read, for the mappers that gate other
// attributes on a toggle rather than mapping it — elasticsearch only exposes
// kibana and apm credentials when those services are on.
func (f Features) Enabled(key string) bool {
	return f[key]
}
