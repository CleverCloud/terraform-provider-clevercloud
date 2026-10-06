package addon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiAddon(slug string) *tmp.AddonResponse {
	return &tmp.AddonResponse{
		ID: "addon_1", Name: "tf-test-addon", Region: "par",
		CreationDate: 1759000000000,
		Plan:         tmp.AddonPlan{Slug: slug},
		Provider:     tmp.AddonResponseProvider{ID: "jenkins"},
	}
}

func TestAddonFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &addon.Addon{}

	state.
		FromAddon(t.Context(), apiAddon("S"), &diags).
		FromEnv(t.Context(), tmp.EnvVars{{Name: "SOME_VAR", Value: "value"}}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"name": state.Name, "plan": state.Plan, "region": state.Region,
		"third_party_provider": state.ThirdPartyProvider,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.CreationDate.IsNull() || state.Configurations.IsNull() {
		t.Error("creation_date and configurations must be set")
	}
}

// Providers disagree on the case of their plan slugs, and the state spelling
// wins when the two differ by case alone — otherwise the plan drifts forever
// (#449, #457).
func TestAddonFromAPI_KeepsTheConfiguredPlanSpelling(t *testing.T) {
	tests := map[string]struct {
		state, api, want string
	}{
		"same case":            {state: "S", api: "S", want: "S"},
		"state keeps its case": {state: "s", api: "S", want: "s"},
		"a real change wins":   {state: "s", api: "m", want: "m"},
		"import takes the API": {state: "", api: "S", want: "S"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics
			state := &addon.Addon{}
			if test.state != "" {
				state.Plan = types.StringValue(test.state)
			}

			state.FromAddon(t.Context(), apiAddon(test.api), &diags)

			if got := state.Plan.ValueString(); got != test.want {
				t.Errorf("plan = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAddonFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &addon.Addon{}
	state.Name = types.StringValue("kept-name")

	state.
		FromAddon(t.Context(), nil, &diags).
		FromEnv(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Name.ValueString() != "kept-name" {
		t.Error("name must be preserved")
	}
}
