package registry

import (
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/resources"
)

// TestAllCodecContracts validates every codec registered via resources.RegisterCodec.
// This package imports all resource packages, so their init() functions have run by
// the time the test executes and the registry is fully populated. A codec/struct
// mismatch (renamed or wrong-typed StateField) is therefore caught here at CI time,
// without a per-resource test file.
func TestAllCodecContracts(t *testing.T) {
	regs := resources.RegisteredCodecs()
	if len(regs) == 0 {
		t.Fatal("no codecs registered; resource init() functions did not run")
	}

	for _, r := range regs {
		t.Run(r.Name, func(t *testing.T) {
			if diags := r.Codec.Validate(r.State); diags.HasError() {
				t.Fatalf("codec/struct mismatch for %q: %v", r.Name, diags)
			}
		})
	}
}
