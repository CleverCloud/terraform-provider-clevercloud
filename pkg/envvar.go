package pkg

import "github.com/miton18/helper/maps"

// EnvVarMapping binds one environment variable to the attribute of S it
// backs. A resource lists its mappings once; writing, reading back and
// rejecting the variables in environment all iterate over that list.
type EnvVarMapping[S any] interface {
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
		name, _ := m.ToEnv(state)
		value := env.PopPtr(name)
		if value == nil {
			m.FromEnv(state, "")
			continue
		}
		m.FromEnv(state, *value)
	}
}

// EnvNames returns the names of the mapped variables. ToEnv must return the
// name whatever the state, empty states included.
func EnvNames[S any](state S, mappings ...EnvVarMapping[S]) []string {
	names := make([]string, 0, len(mappings))
	for _, m := range mappings {
		name, _ := m.ToEnv(state)
		names = append(names, name)
	}
	return names
}
