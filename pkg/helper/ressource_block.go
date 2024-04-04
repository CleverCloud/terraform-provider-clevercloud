package helper

import "strconv"

// Ressource structure

// addon
// resource "clevercloud_addon" "%s" {
// 	name = "%s"
// 	third_party_provider = "mailpace"
// 	plan = "clever_solo"
// 	region = "par"
// }

// cellar
// resource "clevercloud_cellar" "%s" {
// 	name = "%s"
// 	region = "par"
// }

// cellar bucket
// resource "clevercloud_cellar_bucket" "%s" {
//  id = "%s"
//  cellar_id = "%s"
// }

// postgresql
// resource "clevercloud_postgresql" "%s" {
// 	name = "%s"
// 	plan = "dev"
// 	region = "par"
// }

// nodejs
// resource "clevercloud_nodejs" "%s" {
// 	name = "%s"
// 	region = "par"
// 	min_instance_count = 1
// 	max_instance_count = 2
// 	smallest_flavor = "XS"
// 	biggest_flavor = "M"
// 	redirect_https = true
// 	sticky_sessions = true
// 	app_folder = "./app"
// 	environment = {
// 		MY_KEY = "myval"
// 	}
// 	hooks {
// 		post_build = "echo \"build is OK!\""
// 	}
// 	dependencies = []
// }

// resource "clevercloud_nodejs" "%s" {
// 	name = "%s"
// 	region = "par"
// 	min_instance_count = 1
// 	max_instance_count = 2
// 	smallest_flavor = "XS"
// 	biggest_flavor = "M"
// 	deployment {
// 		repository = "https://github.com/CleverCloud/nodejs-example.git"
// 	}
// }

// php
// resource "clevercloud_php" "%s" {
// 	name = "%s"
// 	region = "par"
// 	min_instance_count = 1
// 	max_instance_count = 2
// 	smallest_flavor = "XS"
// 	biggest_flavor = "M"
//  php_version = "8"
// 	additional_vhosts = [ "toto-tf5283457829345.com" ]
// }

// python
// resource "clevercloud_python" "%s" {
// 	name = "%s"
// 	region = "par"
// 	min_instance_count = 1
// 	max_instance_count = 2
// 	smallest_flavor = "XS"
// 	biggest_flavor = "M"
// 	redirect_https = true
// 	sticky_sessions = true
// 	app_folder = "./app"
// 	python_version = "2.7"
// 	pip_requirements = "requirements.txt"
// 	environment = {
// 		MY_KEY = "myval"
// 	}
// 	hooks {
// 		post_build = "echo \"build is OK!\""
// 	}
// 	dependencies = []
// }

// resource "clevercloud_python" "%s" {
// 	name = "%s"
// 	region = "par"
// 	min_instance_count = 1
// 	max_instance_count = 2
// 	smallest_flavor = "XS"
// 	biggest_flavor = "M"
// 	deployment {
// 		repository = "https://github.com/CleverCloud/flask-example.git"
// 	}
// }

// scala
// resource "clevercloud_scala" "%s" {
// 	name = "%s"
// 	region = "par"
// 	min_instance_count = 1
// 	max_instance_count = 2
// 	smallest_flavor = "XS"
// 	biggest_flavor = "M"
// }

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
		sstring := ``
		for key, value := range p.StringValues {
			sstring += `	` + key + ` = "` + value + `"
`
		}
		s += sstring
	}
	// check IntValues not empty
	if len(p.IntValues) != 0 {
		sint := ``
		for key, value := range p.IntValues {
			sint += `	` + key + ` = ` + strconv.Itoa(value) + `
`
		}
		s += sint
	}
	// close s
	s += `}`
	return s
}
