package helper

import (
	"sort"
	"strconv"
)

type Ressource struct {
	Ressource    string
	Name         string
	StringValues map[string]string
	IntValues    map[string]int
}

// New function type that accepts pointer to Ressource
// (~= Signature of option functions)
type RessourceOption func(*Ressource)

// Ressource constructor:
//   - desc: Build a new Ressource and apply specifics RessourceOption functions
//   - args: Ressource name, RessourceOption function
//   - return: pointer to Ressource
func NewRessource(ressource string, opts ...RessourceOption) *Ressource {
	// default values
	const (
		defaultName           = ""
		defaultRegion         = "par"
		dafaultMinInstances   = 1
		defaultMaxInstances   = 2
		defaultSmallestFlavor = "XS"
		defaultBiggestFlavor  = "M"
	)

	var r Ressource
	r.Ressource = ressource
	r.Name = defaultName
	r.StringValues = map[string]string{
		"region":          defaultRegion,
		"smallest_flavor": defaultSmallestFlavor,
		"biggest_flavor":  defaultBiggestFlavor,
	}
	r.IntValues = map[string]int{
		"min_instance_count": dafaultMinInstances,
		"max_instance_count": defaultMaxInstances,
	}

	// RessourceOption functions
	for _, opt := range opts {
		opt(&r)
	}

	return &r
}

// Name value setter:
//   - desc: concatenate function that set Ressource.Name then return Ressource
//   - args: new name
//   - return: pointer to Ressource
func (r *Ressource) SetName(newName string) *Ressource {
	r.Name = newName
	return r
}

// String values setter:
//   - desc: set/add key: value to the string values map of a Ressource then return the Ressource
//   - args: key + value
//   - return: pointer to Ressource
func (p *Ressource) SetStringValues(key, value string) *Ressource {
	p.StringValues[key] = value
	return p
}

// Integer values setter:
//   - desc: set/add key: value to the int values map of a Ressource then return the Ressource
//   - args: key + value
//   - return: pointer to Ressource
func (p *Ressource) SetIntValues(key string, value int) *Ressource {
	p.IntValues[key] = value
	return p
}

// Ressource block
//   - desc: chained function that stringify Ressource into a terraform block
//   - args: none
//   - return: string
func (p *Ressource) String() string {
	s := `ressource "` + p.Ressource + `" "` + p.Name + `" {
	name = "` + p.Name + `"
`
	// check StringValues not empty
	if len(p.StringValues) != 0 {
		// sort StringValues keys
		tmp := make([]string, 0, len(p.StringValues))
		for k := range p.StringValues {
			tmp = append(tmp, k)
		}
		sort.Strings(tmp)

		// create StringValues block
		sstring := ``
		for _, k := range tmp {
			sstring += `	` + k + ` = "` + p.StringValues[k] + `"
`
		}
		s += sstring
	}
	// check IntValues not empty
	if len(p.IntValues) != 0 {
		// sort IntValues keys
		tmp := make([]string, 0, len(p.IntValues))
		for k := range p.IntValues {
			tmp = append(tmp, k)
		}
		sort.Strings(tmp)

		// create IntValues block
		sint := ``
		for _, k := range tmp {
			sint += `	` + k + ` = ` + strconv.Itoa(p.IntValues[k]) + `
`
		}
		s += sint
	}
	// close s
	s += `}`
	return s
}
