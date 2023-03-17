package provider

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

var VhostCleverAppsReg = regexp.MustCompile(`^app-.*\.cleverapps\.io$`)

func IfIsSet(v types.String, fn func(s string)) {
	if !v.IsNull() && !v.IsUnknown() {
		fn(v.ValueString())
	}
}

func IfIsSetB(v types.Bool, fn func(s bool)) {
	if !v.IsNull() && !v.IsUnknown() {
		fn(v.ValueBool())
	}
}

func Map[T, U any](items []T, fn func(T) U) []U {
	r := make([]U, len(items))

	for i, item := range items {
		r[i] = fn(item)
	}

	return r
}

// Filter items in a list
// return true => keep item
func Filter[T any](items []T, fn func(T) bool) []T {
	r := []T{}

	for _, item := range items {
		if fn(item) {
			r = append(r, item)
		}
	}

	return r
}

// Test if any item match a criteria
func HasSome[T any](items []T, fn func(item T) bool) bool {
	for _, item := range items {
		if fn(item) {
			return true
		}
	}

	return false
}

// return the first element mathing a criteria
func First[T any](items []T, fn func(item T) bool) *T {
	for _, item := range items {
		if fn(item) {
			return &item
		}
	}

	return nil
}
