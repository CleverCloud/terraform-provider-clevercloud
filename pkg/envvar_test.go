package pkg

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
)

type testState struct{ a, b types.String }

var testMappings = []EnvVarMapping[*testState]{
	NewStringEnv[testState]("VAR_A",
		func(s *testState) types.String { return s.a },
		func(s *testState, v types.String) { s.a = v }),
	NewStringEnv[testState]("VAR_B",
		func(s *testState) types.String { return s.b },
		func(s *testState, v types.String) { s.b = v }),
}

func TestWriteEnv(t *testing.T) {
	env := map[string]string{"OTHER": "kept"}
	WriteEnv(env, &testState{a: types.StringValue("1"), b: types.StringNull()}, testMappings...)

	if env["VAR_A"] != "1" || env["OTHER"] != "kept" {
		t.Errorf("env = %v", env)
	}
	if _, ok := env["VAR_B"]; ok {
		t.Error("an unset attribute must not be written")
	}
}

func TestReadEnv(t *testing.T) {
	env := maps.NewMap(map[string]string{"VAR_A": "1", "OTHER": "kept"})
	state := &testState{a: types.StringValue("old"), b: types.StringValue("old")}
	ReadEnv(env, state, testMappings...)

	if state.a.ValueString() != "1" {
		t.Errorf("a = %s, want 1", state.a)
	}
	if !state.b.IsNull() {
		t.Errorf("b = %s, want null for an absent variable", state.b)
	}
	if env.Has("VAR_A") || !env.Has("OTHER") {
		t.Errorf("ReadEnv must pop the mapped variables only, env = %v", env)
	}
}

func TestEnvNames(t *testing.T) {
	if got := EnvNames(testMappings...); !slices.Equal(got, []string{"VAR_A", "VAR_B"}) {
		t.Errorf("got %v", got)
	}
}

func TestBoolEnv(t *testing.T) {
	type flag struct{ on types.Bool }
	m := NewBoolEnv[flag]("VAR",
		func(f *flag) string {
			if f.on.ValueBool() {
				return "install"
			}
			return ""
		},
		func(f *flag, value string) {
			if value == "install" {
				f.on = types.BoolValue(true)
			}
		})

	for value, want := range map[types.Bool]string{types.BoolValue(true): "install", types.BoolValue(false): "", types.BoolNull(): ""} {
		if name, got := m.ToEnv(&flag{on: value}); name != "VAR" || got != want {
			t.Errorf("ToEnv(%s) = %s, %q, want VAR, %q", value, name, got, want)
		}
	}

	f := &flag{on: types.BoolValue(false)}
	m.FromEnv(f, "install")
	if !f.on.ValueBool() {
		t.Error(`FromEnv("install") should set true`)
	}
	f = &flag{on: types.BoolValue(false)}
	m.FromEnv(f, "")
	if f.on != types.BoolValue(false) {
		t.Errorf("an absent variable must leave the attribute as is, got %s", f.on)
	}
}
