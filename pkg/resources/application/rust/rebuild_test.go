package rust_test

import (
	"slices"
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/rust"
)

func TestRustRebuildsOnFeaturesChange(t *testing.T) {
	var plan any = &rust.Rust{}
	rt, ok := plan.(application.RebuildTrigger)
	if !ok {
		t.Fatal("rust plan must implement application.RebuildTrigger")
	}
	if got := rt.RebuildEnvKeys(); !slices.Equal(got, []string{rust.CC_RUST_FEATURES}) {
		t.Errorf("RebuildEnvKeys() = %v, want [%s]", got, rust.CC_RUST_FEATURES)
	}
}
