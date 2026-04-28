package networkgroup

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.dev/sdk/models"
)

func strPtr(s string) *string { return &s }

func setNull() basetypes.SetValue {
	return basetypes.NewSetNull(types.StringType)
}

func setOf(t *testing.T, items ...string) basetypes.SetValue {
	t.Helper()
	elems := make([]attr.Value, len(items))
	for i, item := range items {
		elems[i] = basetypes.NewStringValue(item)
	}
	out, d := basetypes.NewSetValue(types.StringType, elems)
	if d.HasError() {
		t.Fatalf("failed to build set: %v", d)
	}
	return out
}

func TestReadFromAPI_Description(t *testing.T) {
	cases := []struct {
		name      string
		stateDesc basetypes.StringValue
		apiDesc   *string
		wantNull  bool
		wantValue string
	}{{
		name:      "null state + nil API → keep null",
		stateDesc: basetypes.NewStringNull(),
		apiDesc:   nil,
		wantNull:  true,
	}, {
		name:      "null state + empty API → keep null",
		stateDesc: basetypes.NewStringNull(),
		apiDesc:   strPtr(""),
		wantNull:  true,
	}, {
		name:      "null state + value API → sync from API",
		stateDesc: basetypes.NewStringNull(),
		apiDesc:   strPtr("hello"),
		wantValue: "hello",
	}, {
		name:      "set state + value API → sync from API",
		stateDesc: basetypes.NewStringValue("old"),
		apiDesc:   strPtr("new"),
		wantValue: "new",
	}, {
		name:      "set state + empty API → sync to empty",
		stateDesc: basetypes.NewStringValue("old"),
		apiDesc:   strPtr(""),
		wantValue: "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &Networkgroup{Description: tc.stateDesc, Tags: setNull()}
			ng := &models.NetworkGroup1{
				Label:       "n",
				NetworkIP:   "10.0.0.0/24",
				Description: tc.apiDesc,
			}
			var diags diag.Diagnostics

			readFromAPI(state, ng, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if tc.wantNull {
				if !state.Description.IsNull() {
					t.Fatalf("expected description to be null, got %q", state.Description.ValueString())
				}
				return
			}
			if state.Description.IsNull() {
				t.Fatalf("expected description %q, got null", tc.wantValue)
			}
			if got := state.Description.ValueString(); got != tc.wantValue {
				t.Fatalf("expected description %q, got %q", tc.wantValue, got)
			}
		})
	}
}

func TestReadFromAPI_Tags(t *testing.T) {
	cases := []struct {
		name      string
		stateTags basetypes.SetValue
		apiTags   []string
		wantNull  bool
		wantTags  []string
	}{{
		name:      "null state + nil API → keep null",
		stateTags: setNull(),
		apiTags:   nil,
		wantNull:  true,
	}, {
		name:      "null state + empty API → keep null",
		stateTags: setNull(),
		apiTags:   []string{},
		wantNull:  true,
	}, {
		name:      "null state + values API → sync from API",
		stateTags: setNull(),
		apiTags:   []string{"a", "b"},
		wantTags:  []string{"a", "b"},
	}, {
		name:      "set state + values API → sync from API",
		stateTags: setOf(t, "old"),
		apiTags:   []string{"new"},
		wantTags:  []string{"new"},
	}, {
		name:      "set state + empty API → sync to empty",
		stateTags: setOf(t, "old"),
		apiTags:   []string{},
		wantTags:  []string{},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &Networkgroup{Description: basetypes.NewStringNull(), Tags: tc.stateTags}
			ng := &models.NetworkGroup1{
				Label:     "n",
				NetworkIP: "10.0.0.0/24",
				Tags:      tc.apiTags,
			}
			var diags diag.Diagnostics

			readFromAPI(state, ng, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if tc.wantNull {
				if !state.Tags.IsNull() {
					t.Fatalf("expected tags to be null, got %v", state.Tags)
				}
				return
			}
			if state.Tags.IsNull() {
				t.Fatalf("expected tags %v, got null", tc.wantTags)
			}

			var got []string
			diags2 := state.Tags.ElementsAs(t.Context(), &got, false)
			if diags2.HasError() {
				t.Fatalf("failed to extract tags: %v", diags2)
			}
			if len(got) != len(tc.wantTags) {
				t.Fatalf("expected tags %v, got %v", tc.wantTags, got)
			}
			// Sets are unordered: check membership
			seen := map[string]bool{}
			for _, v := range got {
				seen[v] = true
			}
			for _, want := range tc.wantTags {
				if !seen[want] {
					t.Fatalf("missing tag %q in result %v", want, got)
				}
			}
		})
	}
}

func TestReadFromAPI_RequiredFields(t *testing.T) {
	state := &Networkgroup{Description: basetypes.NewStringNull(), Tags: setNull()}
	ng := &models.NetworkGroup1{Label: "myng", NetworkIP: "10.42.0.0/16"}
	var diags diag.Diagnostics

	readFromAPI(state, ng, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := state.Name.ValueString(); got != "myng" {
		t.Fatalf("expected name %q, got %q", "myng", got)
	}
	if got := state.Network.ValueString(); got != "10.42.0.0/16" {
		t.Fatalf("expected network %q, got %q", "10.42.0.0/16", got)
	}
}

func TestReadFromAPI_NilSafe(t *testing.T) {
	var diags diag.Diagnostics
	// Should not panic.
	readFromAPI(nil, &models.NetworkGroup1{}, &diags)
	readFromAPI(&Networkgroup{}, nil, &diags)
}
