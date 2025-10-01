package tests

import "fmt"

// assertError helper function for cleaner error assertions
func AssertError(message string, got, want any) error {
	return fmt.Errorf("%s: got %v, want %v", message, got, want)
}
