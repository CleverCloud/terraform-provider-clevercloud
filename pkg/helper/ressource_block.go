package helper

import (
	"maps"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.clever-cloud.com/terraform-provider/pkg"
)

type Ressource struct {
	ressourceType string
	ressourceName string
	keyValues     map[string]any
	blockValues   map[string]any
	isData        bool // true for data sources, false for resources
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
	r.isData = false

	// RessourceOption functions
	for _, opt := range opts {
		opt(&r)
	}

	return &r
}

// DataRessource constructor:
//   - desc: Build a new data source Ressource and apply specifics RessourceOption functions
//   - args: Ressource type and ressource name, RessourceOption function
//   - return: pointer to Ressource
func NewDataRessource(ressourceType, ressourceName string, opts ...RessourceOption) *Ressource {

	var r Ressource
	r.ressourceType = ressourceType
	r.ressourceName = ressourceName
	r.keyValues = map[string]any{}
	r.blockValues = map[string]any{}
	r.isData = true

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

// unit keyValues unset:
//   - desc: remove a key from keyValues field of a Ressource then return the Ressource
//   - args: key to remove
//   - return: pointer to Ressource
func (r *Ressource) UnsetOneValue(key string) *Ressource {
	delete(r.keyValues, key)
	return r
}

// keyValues setter:
//   - desc: set/add key: value to keyValues field of a Ressource then return the Ressource
//   - args: map of string key + value
//   - return: RessourceOption functions
func SetKeyValues(newMap map[string]any) RessourceOption {
	return func(r *Ressource) {
		maps.Copy(r.keyValues, newMap)
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
	var s string
	if r.isData {
		s = `data "` + r.ressourceType + `" "` + r.ressourceName + `" {
`
	} else {
		s = `resource "` + r.ressourceType + `" "` + r.ressourceName + `" {
`
	}

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
		case nil:
			return acc + tab + key + " = null\n"
		case string:
			var_tmp := m[key].(string)
			return acc + tab + key + ` = "` + strings.ReplaceAll(var_tmp, "\"", "\\\"") + `"
`
		case int:
			return acc + tab + key + ` = ` + strconv.Itoa(m[key].(int)) + `
`
		case time.Time:
			return acc + tab + key + ` = "` + m[key].(time.Time).Format(time.RFC3339) + `"
`
		case bool:
			return acc + tab + key + ` = ` + strconv.FormatBool(m[key].(bool)) + `
`
		case map[string]any:
			acc := acc + tab + key + separator + ` {
`
			return map_String(m[key].(map[string]any), acc, `		`, separator) + tab + `}
`
		case []string:
			strs := pkg.Map(c_type, func(s string) string {
				return `"` + s + `"`
			})
			return acc + tab + key + separator + ` [ ` + strings.Join(strs, ", ") + " ]\n"
		case []map[string]string:
			strs := pkg.Map(c_type, func(m map[string]string) string {
				fields := []string{}
				for k, v := range m {
					fields = append(fields, k+` = "`+v+`"`)
				}
				return "{ " + strings.Join(fields, ", ") + " }"
			})
			return acc + tab + key + separator + ` [ ` + strings.Join(strs, ", ") + " ]\n"
		case []map[string]any:
			// Convert each map to a formatted block using existing map_String logic
			blocks := pkg.Map(c_type, func(m map[string]any) string {
				// Build inline block for each map element
				inner := map_String(m, "", "", " =")
				// Remove trailing newline and trim spaces for inline formatting
				inner = strings.TrimSpace(strings.ReplaceAll(inner, "\n", ", "))
				return "{ " + inner + " }"
			})
			return acc + tab + key + separator + ` [ ` + strings.Join(blocks, ", ") + " ]\n"
		default:
			return acc + `// Type ` + reflect.TypeOf(c_type).String() + ` of key "` + key + `" not considered yet
`
		}
	})

	return s
}
