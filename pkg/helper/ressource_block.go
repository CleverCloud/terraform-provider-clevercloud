package helper

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg"
)

type Ressource struct {
	ressourceType string
	ressourceName string
	keyValues     map[string]any
	blockValues   map[string]any
}

// New function type that accepts pointer to Ressource
// (~= Signature of option functions)
type RessourceOption func(*Ressource)

// Ressource constructor:
//   - desc: Build a new Ressource and apply specifics RessourceOption functions
//   - args: Ressource type and ressource name, RessourceOption function
//   - return: pointer to Ressource
func NewRessource(ressourceType, ressourceName string, opts ...RessourceOption) *Ressource {

	var r Ressource
	r.ressourceType = ressourceType
	r.ressourceName = ressourceName
	r.keyValues = map[string]any{}
	r.blockValues = map[string]any{}

	// RessourceOption functions
	for _, opt := range opts {
		opt(&r)
	}

	return &r
}

// unit keyValues setter:
//   - desc: set/add only one key: value to keyvalues field of a Ressource then return the Ressource
//   - args: key + value
//   - return: pointer to Ressource
func (r *Ressource) SetOneValue(key string, value any) *Ressource {
	r.keyValues[key] = value
	return r
}

// keyValues setter:
//   - desc: set/add key: value to keyValues field of a Ressource then return the Ressource
//   - args: map of string key + value
//   - return: RessourceOption functions
func SetKeyValues(newMap map[string]any) RessourceOption {
	return func(r *Ressource) {
		for key, value := range newMap {
			r.keyValues[key] = value
		}
	}
}

// blockValues setter:
//   - desc: set/add key: value to kblockValues field of a Ressource then return the Ressource
//   - args: map of string key + value
//   - return: RessourceOption functions
func SetBlockValues(blockName string, newMap map[string]any) RessourceOption {
	return func(r *Ressource) {
		r.blockValues[blockName] = newMap
	}
}

// Ressource block
//   - desc: chained function that stringify Ressource into a terraform block
//   - args: none
//   - return: string
func (r *Ressource) String() string {
	s := `resource "` + r.ressourceType + `" "` + r.ressourceName + `" {
`

	// create keyValues block
	s = map_String(r.keyValues, s, `	`, ` =`)
	// create blockValues block
	s = map_String(r.blockValues, s, `	`, ``)

	// close s
	s += `}
`

	return s
}

func map_String(m map[string]any, s, tab, separator string) string {
	// sort keyValues keys
	valuesKeys := make([]string, 0, len(m))
	for k := range m {
		valuesKeys = append(valuesKeys, k)
	}
	sort.Strings(valuesKeys)

	// create keyValues block
	s = pkg.Reduce(valuesKeys, s, func(acc, key string) string {
		switch c_type := m[key].(type) {
		case string:
			var_tmp := m[key].(string)
			return acc + tab + key + ` = "` + strings.ReplaceAll(var_tmp, "\"", "\\\"") + `"
`
		case int:
			return acc + tab + key + ` = ` + strconv.Itoa(m[key].(int)) + `
`
		case bool:
			return acc + tab + key + ` = ` + strconv.FormatBool(m[key].(bool)) + `
`
		case map[string]any:
			acc := acc + tab + key + separator + ` {
`
			return map_String(m[key].(map[string]any), acc, `		`, separator) + tab + `}
`
		default:
			return acc + `// Type ` + reflect.TypeOf(c_type).String() + ` of key "` + key + `" not considered yet
`
		}
	})

	return s
}
