package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
)

type sampleState struct {
	Encryption types.Bool
	Kibana     types.Bool
	Locale     types.String
	Region     types.String
	MinCount   types.Int64
}

var sampleCodec = resources.Codec{
	{StateField: "Encryption", APIKeyName: "encryption", Kind: resources.KindBool},
	{StateField: "Kibana", APIKeyName: "kibana", Kind: resources.KindBool},
	{StateField: "Locale", APIKeyName: "locale", Kind: resources.KindString},
	{StateField: "Region", APIKeyName: "region", Kind: resources.KindString},
	{StateField: "MinCount", APIKeyName: "min_count", Kind: resources.KindInt64},
}

func TestStateToAPI_NominalAllSet(t *testing.T) {
	state := sampleState{
		Encryption: types.BoolValue(true),
		Kibana:     types.BoolValue(false),
		Locale:     types.StringValue("en_GB"),
		Region:     types.StringValue("par"),
		MinCount:   types.Int64Value(3),
	}
	api := map[string]any{}

	if diags := sampleCodec.StateToAPI(&state, api); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	want := map[string]any{
		"encryption": true,
		"kibana":     false,
		"locale":     "en_GB",
		"region":     "par",
		"min_count":  int64(3),
	}
	if len(api) != len(want) {
		t.Fatalf("got %d keys, want %d (api=%v)", len(api), len(want), api)
	}
	for k, v := range want {
		if api[k] != v {
			t.Errorf("api[%q]=%v, want %v", k, api[k], v)
		}
	}
}

func TestStateToAPI_NullAndUnknownAreSkipped(t *testing.T) {
	state := sampleState{
		Encryption: types.BoolValue(true),
		Kibana:     types.BoolNull(),
		Locale:     types.StringUnknown(),
		Region:     types.StringValue("par"),
	}
	api := map[string]any{}

	if diags := sampleCodec.StateToAPI(&state, api); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if _, ok := api["kibana"]; ok {
		t.Errorf("null state field should be skipped, got api[kibana]=%v", api["kibana"])
	}
	if _, ok := api["locale"]; ok {
		t.Errorf("unknown state field should be skipped, got api[locale]=%v", api["locale"])
	}
	if api["encryption"] != true {
		t.Errorf("encryption=%v, want true", api["encryption"])
	}
	if api["region"] != "par" {
		t.Errorf("region=%v, want \"par\"", api["region"])
	}
}

func TestAPIToState_NominalAllPresent(t *testing.T) {
	api := map[string]any{
		"encryption": true,
		"kibana":     false,
		"locale":     "en_GB",
		"region":     "par",
		"min_count":  int64(5),
	}
	var state sampleState

	if diags := sampleCodec.APIToState(api, &state); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.Encryption.Equal(types.BoolValue(true)) {
		t.Errorf("Encryption=%v, want true", state.Encryption)
	}
	if !state.Kibana.Equal(types.BoolValue(false)) {
		t.Errorf("Kibana=%v, want false", state.Kibana)
	}
	if !state.Locale.Equal(types.StringValue("en_GB")) {
		t.Errorf("Locale=%v, want \"en_GB\"", state.Locale)
	}
	if !state.Region.Equal(types.StringValue("par")) {
		t.Errorf("Region=%v, want \"par\"", state.Region)
	}
	if !state.MinCount.Equal(types.Int64Value(5)) {
		t.Errorf("MinCount=%v, want 5", state.MinCount)
	}
}

func TestAPIToState_MissingKeys(t *testing.T) {
	api := map[string]any{
		"encryption": true,
		// kibana, locale, region absent
	}
	state := sampleState{
		// Pre-populate to verify absent-key handling per kind.
		Kibana: types.BoolValue(true),
		Locale: types.StringValue("zzz"),
		Region: types.StringValue("zzz"),
	}

	if diags := sampleCodec.APIToState(api, &state); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.Encryption.Equal(types.BoolValue(true)) {
		t.Errorf("Encryption=%v, want true", state.Encryption)
	}
	// KindBool: absent key preserves the existing value (sparse feature list).
	if !state.Kibana.Equal(types.BoolValue(true)) {
		t.Errorf("absent KindBool should preserve prior value, got Kibana=%v", state.Kibana)
	}
	// KindString: absent key yields null.
	if !state.Locale.IsNull() {
		t.Errorf("missing api key should set null, got Locale=%v", state.Locale)
	}
	if !state.Region.IsNull() {
		t.Errorf("missing api key should set null, got Region=%v", state.Region)
	}
}

func TestNonPointerState(t *testing.T) {
	t.Run("StateToAPI accepts a value (read-only)", func(t *testing.T) {
		api := map[string]any{}
		diags := sampleCodec.StateToAPI(sampleState{Encryption: types.BoolValue(true)}, api)
		if diags.HasError() {
			t.Fatalf("value state should encode without error: %v", diags)
		}
		if api["encryption"] != true {
			t.Errorf("encryption=%v, want true", api["encryption"])
		}
	})
	t.Run("APIToState rejects a value (needs pointer to mutate)", func(t *testing.T) {
		diags := sampleCodec.APIToState(map[string]any{}, sampleState{})
		if !diags.HasError() {
			t.Fatal("expected error diagnostic for non-pointer state")
		}
	})
}

func TestStateToAPI_UnknownField(t *testing.T) {
	bad := resources.Codec{
		{StateField: "Nonexistent", APIKeyName: "x", Kind: resources.KindBool},
	}
	diags := bad.StateToAPI(&sampleState{}, map[string]any{})
	if !diags.HasError() {
		t.Fatal("expected error diagnostic for unknown state field")
	}
}

func TestAPIToState_TypeMismatch(t *testing.T) {
	api := map[string]any{
		"encryption": "yes", // wrong type: string instead of bool
	}
	var state sampleState
	diags := sampleCodec.APIToState(api, &state)
	if !diags.HasError() {
		t.Fatal("expected error diagnostic for type mismatch")
	}
}

func TestEnvFlag_Encode(t *testing.T) {
	type s struct{ DevDeps types.Bool }
	c := resources.Codec{
		{StateField: "DevDeps", APIKeyName: "CC_PHP_DEV_DEPENDENCIES", Kind: resources.KindEnvFlag, TruthyValue: "install"},
	}

	t.Run("true writes TruthyValue", func(t *testing.T) {
		api := map[string]any{}
		c.StateToAPI(&s{DevDeps: types.BoolValue(true)}, api)
		if api["CC_PHP_DEV_DEPENDENCIES"] != "install" {
			t.Errorf("got %v, want \"install\"", api["CC_PHP_DEV_DEPENDENCIES"])
		}
	})
	t.Run("false skipped", func(t *testing.T) {
		api := map[string]any{}
		c.StateToAPI(&s{DevDeps: types.BoolValue(false)}, api)
		if _, ok := api["CC_PHP_DEV_DEPENDENCIES"]; ok {
			t.Errorf("false should not write env var, got %v", api["CC_PHP_DEV_DEPENDENCIES"])
		}
	})
	t.Run("null skipped", func(t *testing.T) {
		api := map[string]any{}
		c.StateToAPI(&s{DevDeps: types.BoolNull()}, api)
		if _, ok := api["CC_PHP_DEV_DEPENDENCIES"]; ok {
			t.Errorf("null should not write env var")
		}
	})
}

func TestEnvFlag_Decode(t *testing.T) {
	type s struct{ DevDeps types.Bool }
	c := resources.Codec{
		{StateField: "DevDeps", APIKeyName: "CC_PHP_DEV_DEPENDENCIES", Kind: resources.KindEnvFlag, TruthyValue: "install"},
	}

	t.Run("matching value yields true", func(t *testing.T) {
		st := s{DevDeps: types.BoolNull()}
		c.APIToState(map[string]any{"CC_PHP_DEV_DEPENDENCIES": "install"}, &st)
		if !st.DevDeps.Equal(types.BoolValue(true)) {
			t.Errorf("DevDeps=%v, want true", st.DevDeps)
		}
	})
	t.Run("non-matching value preserves prior", func(t *testing.T) {
		st := s{DevDeps: types.BoolValue(false)}
		c.APIToState(map[string]any{"CC_PHP_DEV_DEPENDENCIES": "ignore"}, &st)
		if !st.DevDeps.Equal(types.BoolValue(false)) {
			t.Errorf("non-match should preserve prior, got %v", st.DevDeps)
		}
	})
	t.Run("absent preserves prior (null stays null)", func(t *testing.T) {
		st := s{DevDeps: types.BoolNull()}
		c.APIToState(map[string]any{}, &st)
		if !st.DevDeps.IsNull() {
			t.Errorf("absent should preserve null, got %v", st.DevDeps)
		}
	})
}

func TestBoolString(t *testing.T) {
	type s struct{ Mount types.Bool }
	c := resources.Codec{
		{StateField: "Mount", APIKeyName: "CC_MOUNT_DOCKER_SOCKET", Kind: resources.KindBoolString},
	}

	t.Run("encode writes both states", func(t *testing.T) {
		for _, tc := range []struct {
			in   types.Bool
			want any
		}{
			{types.BoolValue(true), "true"},
			{types.BoolValue(false), "false"},
		} {
			api := map[string]any{}
			c.StateToAPI(&s{Mount: tc.in}, api)
			if api["CC_MOUNT_DOCKER_SOCKET"] != tc.want {
				t.Errorf("encode %v => %v, want %v", tc.in, api["CC_MOUNT_DOCKER_SOCKET"], tc.want)
			}
		}
	})
	t.Run("null skipped on encode", func(t *testing.T) {
		api := map[string]any{}
		c.StateToAPI(&s{Mount: types.BoolNull()}, api)
		if _, ok := api["CC_MOUNT_DOCKER_SOCKET"]; ok {
			t.Error("null should not be written")
		}
	})
	t.Run("decode parses, absent and invalid yield null", func(t *testing.T) {
		var st s
		c.APIToState(map[string]any{"CC_MOUNT_DOCKER_SOCKET": "true"}, &st)
		if !st.Mount.Equal(types.BoolValue(true)) {
			t.Errorf("decode \"true\" => %v", st.Mount)
		}
		st = s{}
		c.APIToState(map[string]any{}, &st)
		if !st.Mount.IsNull() {
			t.Errorf("absent should be null, got %v", st.Mount)
		}
		st = s{}
		c.APIToState(map[string]any{"CC_MOUNT_DOCKER_SOCKET": "nope"}, &st)
		if !st.Mount.IsNull() {
			t.Errorf("unparseable should be null, got %v", st.Mount)
		}
	})
}

func TestValidate(t *testing.T) {
	if diags := sampleCodec.Validate(&sampleState{}); diags.HasError() {
		t.Fatalf("valid codec should pass: %v", diags)
	}
	missing := resources.Codec{{StateField: "Nope", APIKeyName: "x", Kind: resources.KindBool}}
	if !missing.Validate(&sampleState{}).HasError() {
		t.Error("expected error for missing field")
	}
	wrongType := resources.Codec{{StateField: "Locale", APIKeyName: "locale", Kind: resources.KindBool}}
	if !wrongType.Validate(&sampleState{}).HasError() {
		t.Error("expected error for Kind/field-type mismatch")
	}
}

func TestAPIToState_FieldTypeMismatch(t *testing.T) {
	type badState struct {
		Encryption types.String // wrong: declared as String in struct but Bool in Codec
	}
	c := resources.Codec{
		{StateField: "Encryption", APIKeyName: "encryption", Kind: resources.KindBool},
	}
	diags := c.StateToAPI(&badState{Encryption: types.StringValue("nope")}, map[string]any{})
	if !diags.HasError() {
		t.Fatal("expected error diagnostic for field-type / Kind mismatch")
	}
}
