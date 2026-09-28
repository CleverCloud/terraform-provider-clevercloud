package pkg_test

import (
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg"
)

func TestBypassSensitiveImports(t *testing.T) {
	tests := map[string]struct {
		value string
		want  bool
	}{
		"unset":       {value: "", want: false},
		"true":        {value: "true", want: true},
		"True":        {value: "True", want: true},
		"TRUE":        {value: "TRUE", want: true},
		"yes":         {value: "yes", want: true},
		"YES":         {value: "YES", want: true},
		"1":           {value: "1", want: true},
		"t":           {value: "t", want: true},
		"padded":      {value: "  true  ", want: true},
		"false":       {value: "false", want: false},
		"no":          {value: "no", want: false},
		"0":           {value: "0", want: false},
		"f":           {value: "f", want: false},
		"unparseable": {value: "maybe", want: false},
		"nearly":      {value: "truthy", want: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(pkg.BypassSensitiveImportsEnvVar, test.value)

			if got := pkg.BypassSensitiveImports(); got != test.want {
				t.Errorf("with %q, got %t, want %t", test.value, got, test.want)
			}
		})
	}
}
