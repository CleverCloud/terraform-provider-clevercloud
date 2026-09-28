package pkg_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

func strPtr(s string) *string { return &s }

// SetBoolIf leaves the attribute alone when the variable does not match, which
// is what an attribute with no schema default wants: null keeps meaning "not
// configured".
func TestSetBoolIf(t *testing.T) {
	tests := map[string]struct {
		value *string
		want  types.Bool
	}{
		"match":    {value: strPtr("true"), want: types.BoolValue(true)},
		"mismatch": {value: strPtr("false"), want: types.BoolNull()},
		"absent":   {value: nil, want: types.BoolNull()},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			target := types.BoolNull()
			pkg.SetBoolIf(&target, test.value, "true")

			if !target.Equal(test.want) {
				t.Errorf("got %v, want %v", target, test.want)
			}
		})
	}
}

// SetBool always assigns, so an attribute carrying a schema default is never
// left null after an import — which is what made the next plan non-empty.
func TestSetBool(t *testing.T) {
	tests := map[string]struct {
		value *string
		want  types.Bool
	}{
		"match":    {value: strPtr("true"), want: types.BoolValue(true)},
		"mismatch": {value: strPtr("false"), want: types.BoolValue(false)},
		"absent":   {value: nil, want: types.BoolValue(false)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			target := types.BoolNull()
			pkg.SetBool(&target, test.value, "true")

			if !target.Equal(test.want) {
				t.Errorf("got %v, want %v", target, test.want)
			}
		})
	}
}
