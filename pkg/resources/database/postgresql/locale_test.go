package postgresql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// TestReadFromAPI_Locale pins the locale resolution: the API value wins, a
// value already in state is kept when the API does not expose it, and the
// platform default is used on import (no value in state).
func TestReadFromAPI_Locale(t *testing.T) {
	cases := []struct {
		name     string
		api      string
		state    types.String
		expected string
	}{
		{name: "api value overrides state", api: "fr_FR", state: pkg.FromStr("en_GB"), expected: "fr_FR"},
		{name: "api value fills null state", api: "fr_FR", state: types.StringNull(), expected: "fr_FR"},
		{name: "state kept when api is empty", api: "", state: pkg.FromStr("de_DE"), expected: "de_DE"},
		{name: "default when api is empty and state is null", api: "", state: types.StringNull(), expected: defaultLocale},
		{name: "default when api is empty and state is unknown", api: "", state: types.StringUnknown(), expected: defaultLocale},
	}

	r := &ResourcePostgreSQL{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := PostgreSQL{Locale: tc.state}
			r.readFromAPI(t.Context(), &state, tmp.PostgreSQL{Locale: tc.api})
			if got := state.Locale.ValueString(); got != tc.expected {
				t.Errorf("locale = %q, expected %q", got, tc.expected)
			}
		})
	}
}
