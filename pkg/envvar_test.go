package pkg

import (
	"slices"
	"testing"

	"github.com/miton18/helper/maps"
)

type testState struct{ a, b string }

type varA struct{}

func (varA) FromEnv(s *testState, v string)      { s.a = v }
func (varA) ToEnv(s *testState) (string, string) { return "VAR_A", s.a }

type varB struct{}

func (varB) FromEnv(s *testState, v string)      { s.b = v }
func (varB) ToEnv(s *testState) (string, string) { return "VAR_B", s.b }

var testMappings = []EnvVarMapping[*testState]{varA{}, varB{}}

func TestWriteEnv(t *testing.T) {
	env := map[string]string{"OTHER": "kept"}
	WriteEnv(env, &testState{a: "1"}, testMappings...)

	if env["VAR_A"] != "1" || env["OTHER"] != "kept" {
		t.Errorf("env = %v", env)
	}
	if _, ok := env["VAR_B"]; ok {
		t.Error("an unset attribute must not be written")
	}
}

func TestReadEnv(t *testing.T) {
	env := maps.NewMap(map[string]string{"VAR_A": "1", "OTHER": "kept"})
	state := &testState{a: "old", b: "old"}
	ReadEnv(env, state, testMappings...)

	if state.a != "1" {
		t.Errorf("a = %q, want 1", state.a)
	}
	if state.b != "" {
		t.Errorf("b = %q, want \"\" for an absent variable", state.b)
	}
	if env.Has("VAR_A") || !env.Has("OTHER") {
		t.Errorf("ReadEnv must pop the mapped variables only, env = %v", env)
	}
}

func TestEnvNames(t *testing.T) {
	if got := EnvNames(&testState{}, testMappings...); !slices.Equal(got, []string{"VAR_A", "VAR_B"}) {
		t.Errorf("got %v", got)
	}
}
