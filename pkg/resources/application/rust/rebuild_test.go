package rust_test

import (
	"slices"
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application/rust"
	"go.clever-cloud.dev/sdk/app"
)

func TestRustRebuildsOnFeaturesChange(t *testing.T) {
	var plan any = &rust.Rust{}
	rt, ok := plan.(application.RebuildTrigger)
	if !ok {
		t.Fatal("rust plan must implement application.RebuildTrigger")
	}
	if got := rt.RebuildEnvKeys(); !slices.Equal(got, []string{app.CC_RUST_FEATURES}) {
		t.Errorf("RebuildEnvKeys() = %v, want [%s]", got, app.CC_RUST_FEATURES)
	}
}
