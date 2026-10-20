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

// BoolEnv maps a boolean attribute of S to a variable that turns it on when
// it holds trueValue.
type BoolEnv[S any] struct {
	name      string
	trueValue string
	get       func(*S) types.Bool
	set       func(*S, types.Bool)
}

// NewBoolEnv maps the boolean attribute read by get and written by set to the
// variable name, written as trueValue when the attribute is true.
func NewBoolEnv[S any](name, trueValue string, get func(*S) types.Bool, set func(*S, types.Bool)) BoolEnv[S] {
	return BoolEnv[S]{name: name, trueValue: trueValue, get: get, set: set}
}

func (e BoolEnv[S]) Name() string { return e.name }

// FromEnv sets the attribute to true when the variable holds trueValue. Any
// other value, or an absent variable, leaves the attribute as it is.
func (e BoolEnv[S]) FromEnv(s *S, value string) {
	if value == e.trueValue {
		e.set(s, types.BoolValue(true))
	}
}

// ToEnv writes trueValue when the attribute is true, nothing otherwise.
func (e BoolEnv[S]) ToEnv(s *S) (string, string) {
	if e.get(s).ValueBool() {
		return e.name, e.trueValue
	}
	return e.name, ""
}
