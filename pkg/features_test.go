package pkg_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type feature struct {
	Name    string
	Enabled bool
}

func featuresOf(items ...feature) pkg.Features {
	return pkg.FeaturesOf(items, func(f feature) (string, bool) {
		return f.Name, f.Enabled
	})
}

// FeaturesOf keeps what the API said, false included — the distinction that
// makes the absent case decidable at all.
func TestFeaturesOf(t *testing.T) {
	features := featuresOf(
		feature{Name: "do-backup", Enabled: true},
		feature{Name: "encryption", Enabled: false},
	)

	if got := len(features); got != 2 {
		t.Fatalf("got %d features, want 2", got)
	}
	if !features.Enabled("do-backup") {
		t.Error("do-backup should be enabled")
	}
	if features.Enabled("encryption") {
		t.Error("encryption should be disabled")
	}
	if features.Enabled("absent") {
		t.Error("an absent toggle should read as disabled")
	}
}

func TestFeaturesOf_empty(t *testing.T) {
	if got := len(pkg.FeaturesOf(nil, func(f feature) (string, bool) { return f.Name, f.Enabled })); got != 0 {
		t.Fatalf("got %d features from a nil list, want 0", got)
	}
}

// Or always assigns, so an attribute carrying a schema default is never left
// null after an import — the regression #404 and #452 came from.
func TestFeaturesOr(t *testing.T) {
	tests := map[string]struct {
		features pkg.Features
		fallback bool
		want     types.Bool
	}{
		"enabled by the API":          {features: featuresOf(feature{"do-backup", true}), fallback: false, want: types.BoolValue(true)},
		"disabled by the API":         {features: featuresOf(feature{"do-backup", false}), fallback: true, want: types.BoolValue(false)},
		"absent falls back to true":   {features: featuresOf(), fallback: true, want: types.BoolValue(true)},
		"absent falls back to false":  {features: featuresOf(), fallback: false, want: types.BoolValue(false)},
		"absent from a nil list":      {features: nil, fallback: true, want: types.BoolValue(true)},
		"another toggle is no answer": {features: featuresOf(feature{"encryption", true}), fallback: false, want: types.BoolValue(false)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			target := types.BoolNull()
			test.features.Or(&target, "do-backup", test.fallback)

			if !target.Equal(test.want) {
				t.Errorf("got %v, want %v", target, test.want)
			}
		})
	}
}

// Or overwrites a prior state value: the API is the authority on a toggle it
// does report.
func TestFeaturesOr_overwritesPriorState(t *testing.T) {
	target := types.BoolValue(true)
	featuresOf(feature{"do-backup", false}).Or(&target, "do-backup", true)

	if !target.Equal(types.BoolValue(false)) {
		t.Errorf("got %v, want false", target)
	}
}

// Keep leaves the attribute alone when the API is silent, so a null keeps
// meaning "not configured" on an Optional attribute with no default.
func TestFeaturesKeep(t *testing.T) {
	tests := map[string]struct {
		features pkg.Features
		prior    types.Bool
		want     types.Bool
	}{
		"reported, overwrites null":  {features: featuresOf(feature{"kibana", true}), prior: types.BoolNull(), want: types.BoolValue(true)},
		"reported, overwrites prior": {features: featuresOf(feature{"kibana", false}), prior: types.BoolValue(true), want: types.BoolValue(false)},
		"absent, keeps null":         {features: featuresOf(), prior: types.BoolNull(), want: types.BoolNull()},
		"absent, keeps prior":        {features: featuresOf(), prior: types.BoolValue(true), want: types.BoolValue(true)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			target := test.prior
			test.features.Keep(&target, "kibana")

			if !target.Equal(test.want) {
				t.Errorf("got %v, want %v", target, test.want)
			}
		})
	}
}

func TestFeaturesNull(t *testing.T) {
	tests := map[string]struct {
		features pkg.Features
		prior    types.Bool
		want     types.Bool
	}{
		"reported":                   {features: featuresOf(feature{"apm", true}), prior: types.BoolNull(), want: types.BoolValue(true)},
		"absent nulls a prior value": {features: featuresOf(), prior: types.BoolValue(true), want: types.BoolNull()},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			target := test.prior
			test.features.Null(&target, "apm")

			if !target.Equal(test.want) {
				t.Errorf("got %v, want %v", target, test.want)
			}
		})
	}
}
