package application

import (
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	helpermaps "github.com/miton18/helper/maps"
)

func stateEnvMap(t *testing.T, entries map[string]string) types.Map {
	t.Helper()

	values := map[string]attr.Value{}
	for key, value := range entries {
		values[key] = types.StringValue(value)
	}

	return types.MapValueMust(types.StringType, values)
}

func TestWithholdUserManagedEnv(t *testing.T) {
	cases := map[string]struct {
		apiEnv       map[string]string
		stateEnv     types.Map
		wantWithheld map[string]string
		wantLeft     map[string]string
	}{
		"no environment in state": {
			apiEnv:       map[string]string{"CC_RUST_FEATURES": "a,b"},
			stateEnv:     types.MapNull(types.StringType),
			wantWithheld: map[string]string{},
			wantLeft:     map[string]string{"CC_RUST_FEATURES": "a,b"},
		},
		"unknown environment in state": {
			apiEnv:       map[string]string{"CC_RUST_FEATURES": "a,b"},
			stateEnv:     types.MapUnknown(types.StringType),
			wantWithheld: map[string]string{},
			wantLeft:     map[string]string{"CC_RUST_FEATURES": "a,b"},
		},
		"derived var declared in environment": {
			apiEnv:       map[string]string{"CC_RUST_FEATURES": "a,b", "MY_VAR": "42"},
			stateEnv:     stateEnvMap(t, map[string]string{"CC_RUST_FEATURES": "a,b"}),
			wantWithheld: map[string]string{"CC_RUST_FEATURES": "a,b"},
			wantLeft:     map[string]string{"MY_VAR": "42"},
		},
		"derived var declared through its attribute": {
			apiEnv:       map[string]string{"CC_RUST_FEATURES": "a,b"},
			stateEnv:     stateEnvMap(t, map[string]string{"MY_VAR": "42"}),
			wantWithheld: map[string]string{},
			wantLeft:     map[string]string{"CC_RUST_FEATURES": "a,b"},
		},
		"declared var dropped on the API side": {
			apiEnv:       map[string]string{"MY_VAR": "42"},
			stateEnv:     stateEnvMap(t, map[string]string{"CC_PRE_BUILD_HOOK": "make"}),
			wantWithheld: map[string]string{},
			wantLeft:     map[string]string{"MY_VAR": "42"},
		},
		"value changed on the API side": {
			apiEnv:       map[string]string{"CC_PRE_BUILD_HOOK": "make build"},
			stateEnv:     stateEnvMap(t, map[string]string{"CC_PRE_BUILD_HOOK": "make"}),
			wantWithheld: map[string]string{"CC_PRE_BUILD_HOOK": "make build"},
			wantLeft:     map[string]string{},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			env := helpermaps.NewMap(tt.apiEnv)

			withheld := withholdUserManagedEnv(env, tt.stateEnv)

			if !maps.Equal(withheld, tt.wantWithheld) {
				t.Errorf("withheld: expected %v, got %v", tt.wantWithheld, withheld)
			}

			if left := maps.Collect(env.All); !maps.Equal(left, tt.wantLeft) {
				t.Errorf("remaining env: expected %v, got %v", tt.wantLeft, left)
			}
		})
	}
}
