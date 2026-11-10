package tmp

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"
)

// addonprovidersIsDedicated is GET /v2/products/addonproviders, captured on
// 2026-11-11 and trimmed to the providers that expose the "is-dedicated" plan
// feature (es-addon, jenkins, mongodb-addon, mysql-addon, postgresql-addon,
// redis-addon), keeping that feature only.
//
//go:embed testdata/addonproviders_is_dedicated.json
var addonprovidersIsDedicated []byte

func loadCatalogue(t *testing.T) []AddonProvider {
	t.Helper()
	var providers []AddonProvider
	if err := json.Unmarshal(addonprovidersIsDedicated, &providers); err != nil {
		t.Fatalf("cannot decode the catalogue fixture: %v", err)
	}
	if len(providers) == 0 {
		t.Fatal("empty catalogue fixture")
	}
	return providers
}

// Oracle independent from IsDedicated(): Clever Cloud's documentation describes
// the DEV plans as shared and every other database plan as dedicated, which
// holds for all 137 plans of the fixture.
func expectedDedicated(slug string) bool {
	return slug != "dev"
}

// TestIsDedicated_RealCatalogue runs IsDedicated() on every plan of the real
// catalogue. Before reading the feature's value, every MySQL plan was reported
// as shared (its computable_value is "false"/"true", not "dedicated"/"shared").
func TestIsDedicated_RealCatalogue(t *testing.T) {
	plans := 0
	for _, provider := range loadCatalogue(t) {
		for _, plan := range provider.Plans {
			plans++
			t.Run(provider.ID+"/"+plan.Slug, func(t *testing.T) {
				if got, want := plan.IsDedicated(), expectedDedicated(plan.Slug); got != want {
					t.Errorf("IsDedicated() = %v, want %v (feature: %+v)", got, want, plan.Features)
				}
			})
		}
	}
	if plans != 137 {
		t.Errorf("fixture holds %d plans, expected 137", plans)
	}
}

// TestIsDedicated_MySQLDedicatedPlansExist guards the regression this fixes: the
// catalogue does have dedicated MySQL plans, so encryption, skip_log_bin and
// direct_host_only must be accepted on at least one of them.
func TestIsDedicated_MySQLDedicatedPlansExist(t *testing.T) {
	var dedicated []string
	for _, provider := range loadCatalogue(t) {
		if provider.ID != "mysql-addon" {
			continue
		}
		for _, plan := range provider.Plans {
			if plan.IsDedicated() {
				dedicated = append(dedicated, plan.Slug)
			}
		}
	}
	for _, slug := range []string{"xxs_sml", "xs_sml", "m_med"} {
		found := false
		for _, d := range dedicated {
			if d == slug {
				found = true
			}
		}
		if !found {
			t.Errorf("mysql-addon plan %q should be dedicated; dedicated plans found: %v", slug, dedicated)
		}
	}
}

// TestIsDedicated_EncodingsDifferAcrossProviders documents why computable_value
// cannot be trusted: the same feature is encoded two ways in the API, while
// value is consistent.
func TestIsDedicated_EncodingsDifferAcrossProviders(t *testing.T) {
	computable := map[string]map[string]bool{} // provider -> computable values seen on dedicated plans
	for _, provider := range loadCatalogue(t) {
		for _, plan := range provider.Plans {
			for _, f := range plan.Features {
				if f.NameCode != "is-dedicated" {
					continue
				}
				if expectedDedicated(plan.Slug) {
					if !strings.EqualFold(f.Value, "dedicated") {
						t.Errorf("%s/%s: dedicated plan with value %q", provider.ID, plan.Slug, f.Value)
					}
					if computable[provider.ID] == nil {
						computable[provider.ID] = map[string]bool{}
					}
					computable[provider.ID][f.ComputableValue] = true
				} else if !strings.EqualFold(f.Value, "shared") {
					t.Errorf("%s/%s: shared plan with value %q", provider.ID, plan.Slug, f.Value)
				}
			}
		}
	}
	if !computable["mysql-addon"]["false"] || computable["mysql-addon"]["dedicated"] {
		t.Errorf("expected MySQL dedicated plans to carry computable_value \"false\", got %v", computable["mysql-addon"])
	}
	if !computable["postgresql-addon"]["dedicated"] {
		t.Errorf("expected PostgreSQL dedicated plans to carry computable_value \"dedicated\", got %v", computable["postgresql-addon"])
	}
}

// TestAddonPlanFeature_DecodesValue makes sure the "value" field of the API
// response is decoded (it was not part of AddonPlanFeature).
func TestAddonPlanFeature_DecodesValue(t *testing.T) {
	raw := `{"name":"TYPE","type":"BOOLEAN_SHARED","value":"Dedicated","computable_value":"false","name_code":"is-dedicated"}`
	var f AddonPlanFeature
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatal(err)
	}
	if f.Value != "Dedicated" || f.ComputableValue != "false" || f.NameCode != "is-dedicated" {
		t.Errorf("decoded %+v", f)
	}
	plan := AddonPlan{Slug: "m_med", Features: []AddonPlanFeature{f}}
	if !plan.IsDedicated() {
		t.Error("a MySQL m_med plan decoded from the API must be dedicated")
	}
}
