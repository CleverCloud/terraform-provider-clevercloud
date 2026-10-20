package rust

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// featuresEnv maps the features set to CC_RUST_FEATURES, a comma separated
// list.
type featuresEnv struct{}

var _ pkg.EnvVarMapping[*Rust] = featuresEnv{}

func (featuresEnv) Name() string { return app.CC_RUST_FEATURES }

// FromEnv sets features from the list, null when it is empty.
func (featuresEnv) FromEnv(r *Rust, value string) {
	// the elements are strings: building the set cannot fail
	r.Features = pkg.FromSetSplit(&value, ",", &diag.Diagnostics{})
}

func (featuresEnv) ToEnv(r *Rust) (string, string) {
	features := []string{}
	for _, e := range r.Features.Elements() {
		if s, ok := e.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
			features = append(features, s.ValueString())
		}
	}
	return app.CC_RUST_FEATURES, strings.Join(features, ",")
}

var (
	FeaturesEnv      = featuresEnv{}
	RustupChannelEnv = pkg.NewStringEnv[Rust](app.CC_RUSTUP_CHANNEL,
		func(r *Rust) types.String { return r.RustupChannel },
		func(r *Rust, v types.String) { r.RustupChannel = v })
	BinEnv = pkg.NewStringEnv[Rust](app.CC_RUST_BIN,
		func(r *Rust) types.String { return r.Bin },
		func(r *Rust, v types.String) { r.Bin = v })
)

// envVars lists every rust attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*Rust]{
	application.OnRuntime[*Rust](application.AppFolderEnv),
	FeaturesEnv,
	RustupChannelEnv,
	BinEnv,
}
