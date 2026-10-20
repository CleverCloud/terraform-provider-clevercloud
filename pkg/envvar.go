package pkg

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
)

// EnvVarMapping binds one environment variable to the attribute of S it
// backs. A resource lists its mappings once; writing, reading back and
// rejecting the variables in environment all iterate over that list.
type EnvVarMapping[S any] interface {
	// Name returns the variable name.
	Name() string
	// FromEnv sets the attribute from the variable value, "" when the
	// variable is absent.
	FromEnv(state S, value string)
	// ToEnv returns the variable name and the value to write, "" when the
	// attribute is not set.
	ToEnv(state S) (name, value string)
}

// WriteEnv writes into env the variables of the mapped attributes that are
// set.
func WriteEnv[S any](env map[string]string, state S, mappings ...EnvVarMapping[S]) {
	for _, m := range mappings {
		if name, value := m.ToEnv(state); value != "" {
			env[name] = value
		}
	}
}

// ReadEnv moves the mapped variables out of env into the attributes of
// state; an absent variable reaches its mapping as "".
func ReadEnv[S any](env *maps.Map[string, string], state S, mappings ...EnvVarMapping[S]) {
	for _, m := range mappings {
		value := env.PopPtr(m.Name())
		if value == nil {
			// state holds the previous state: FromEnv must still run so an
			// attribute whose variable was removed outside Terraform goes
			// back to null and the drift shows up in the plan.
			m.FromEnv(state, "")
			continue
		}
		m.FromEnv(state, *value)
	}
}

// EnvNames returns the names of the mapped variables.
func EnvNames[S any](mappings ...EnvVarMapping[S]) []string {
	names := make([]string, 0, len(mappings))
	for _, m := range mappings {
		names = append(names, m.Name())
	}
	return names
}

// StringEnv maps a string attribute of S to its environment variable.
type StringEnv[S any] struct {
	name string
	get  func(*S) types.String
	set  func(*S, types.String)
}

// NewStringEnv maps the string attribute read by get and written by set to
// the variable name.
func NewStringEnv[S any](name string, get func(*S) types.String, set func(*S, types.String)) StringEnv[S] {
	return StringEnv[S]{name: name, get: get, set: set}
}

func (e StringEnv[S]) Name() string { return e.name }

// FromEnv sets the attribute to the value, null when it is empty.
func (e StringEnv[S]) FromEnv(s *S, value string) { e.set(s, FromStr(value)) }

func (e StringEnv[S]) ToEnv(s *S) (string, string) { return e.name, e.get(s).ValueString() }

// BoolEnv maps a boolean attribute of S to its variable, through functions
// converting the attribute to the variable value and back.
type BoolEnv[S any] struct {
	name    string
	toEnv   func(*S) string
	fromEnv func(*S, string)
}

// NewBoolEnv maps a boolean attribute to the variable name: toEnv returns the
// value to write ("" to write nothing) and fromEnv sets the attribute from
// the variable value ("" when it is absent).
func NewBoolEnv[S any](name string, toEnv func(*S) string, fromEnv func(*S, string)) BoolEnv[S] {
	return BoolEnv[S]{name: name, toEnv: toEnv, fromEnv: fromEnv}
}

func (e BoolEnv[S]) Name() string                { return e.name }
func (e BoolEnv[S]) FromEnv(s *S, value string)  { e.fromEnv(s, value) }
func (e BoolEnv[S]) ToEnv(s *S) (string, string) { return e.name, e.toEnv(s) }
