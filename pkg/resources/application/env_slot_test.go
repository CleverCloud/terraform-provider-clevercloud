package application

import (
	"context"
	"maps"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	helpermaps "github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
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

// fakeRuntime stands in for a real runtime: CC_FAKE_FEATURE is its only
// attribute-derived variable.
type fakeRuntime struct {
	Runtime
	feature types.String
}

func (f fakeRuntime) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	custom := map[string]string{}
	diags.Append(f.Environment.ElementsAs(ctx, &custom, false)...)
	env = pkg.Merge(env, custom)

	pkg.IfIsSetStr(f.feature, func(value string) { env["CC_FAKE_FEATURE"] = value })

	return env
}

func (f fakeRuntime) ToDeployment(*http.BasicAuth) *Deployment { return nil }

func (f *fakeRuntime) FromEnv(_ context.Context, env *helpermaps.Map[string, string], _ *diag.Diagnostics) {
	f.feature = pkg.FromStrPtr(env.PopPtr("CC_FAKE_FEATURE"))
}

const (
	duplicateSlotSummary = "Variable also managed by a dedicated attribute"
	dedicatedSlotSummary = "Variable has a dedicated attribute"
)

func countBySummary(diags diag.Diagnostics, summary string) int {
	count := 0
	for _, d := range diags.Warnings() {
		if d.Summary() == summary {
			count++
		}
	}

	return count
}

func TestValidateEnvSlots(t *testing.T) {
	cases := map[string]struct {
		environment   types.Map
		feature       types.String
		wantDuplicate int
		wantDedicated int
	}{
		"declared in both slots": {
			environment:   stateEnvMap(t, map[string]string{"CC_FAKE_FEATURE": "a", "MY_VAR": "42"}),
			feature:       types.StringValue("a"),
			wantDuplicate: 1,
		},
		"declared in environment only": {
			environment:   stateEnvMap(t, map[string]string{"CC_FAKE_FEATURE": "a"}),
			feature:       types.StringNull(),
			wantDedicated: 1,
		},
		"declared through the attribute only": {
			environment: stateEnvMap(t, map[string]string{"MY_VAR": "42"}),
			feature:     types.StringValue("a"),
		},
		"variable with no dedicated attribute": {
			environment: stateEnvMap(t, map[string]string{"MY_VAR": "42"}),
			feature:     types.StringNull(),
		},
		"no environment": {
			environment: types.MapNull(types.StringType),
			feature:     types.StringValue("a"),
		},
		"unknown environment": {
			environment: types.MapUnknown(types.StringType),
			feature:     types.StringValue("a"),
		},
		// a value interpolated from a resource not created yet must not blind the check
		"unknown value beside a duplicate": {
			environment: types.MapValueMust(types.StringType, map[string]attr.Value{
				"CC_FAKE_FEATURE": types.StringValue("a"),
				"INTERPOLATED":    types.StringUnknown(),
			}),
			feature:       types.StringValue("a"),
			wantDuplicate: 1,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			plan := &fakeRuntime{feature: tt.feature}
			plan.Environment = tt.environment

			diags := diag.Diagnostics{}
			ValidateEnvSlots(context.Background(), plan, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected errors: %v", diags.Errors())
			}

			if got := countBySummary(diags, duplicateSlotSummary); got != tt.wantDuplicate {
				t.Errorf("expected %d duplicate-slot warning(s), got %d: %v", tt.wantDuplicate, got, diags.Warnings())
			}

			if got := countBySummary(diags, dedicatedSlotSummary); got != tt.wantDedicated {
				t.Errorf("expected %d dedicated-slot warning(s), got %d: %v", tt.wantDedicated, got, diags.Warnings())
			}

			if !plan.Environment.Equal(tt.environment) {
				t.Errorf("environment was not restored: %v", plan.Environment)
			}

			if !plan.feature.Equal(tt.feature) {
				t.Errorf("plan attribute was mutated: %v", plan.feature)
			}
		})
	}
}
