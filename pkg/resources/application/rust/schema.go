package rust

import (
	"context"
	_ "embed"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type Rust struct {
	application.Runtime
	Features      types.Set    `tfsdk:"features"`
	RunCommand    types.String `tfsdk:"run_command"`
	RustBin       types.String `tfsdk:"rust_bin"`
	RustupChannel types.String `tfsdk:"rustup_channel"`
}

func (r Rust) FeaturesAsStrings(ctx context.Context, diags *diag.Diagnostics) []string {
	features := []string{}
	diags.Append(r.Features.ElementsAs(ctx, &features, true)...)
	return features
}

//go:embed doc.md
var rustDoc string

func (r ResourceRust) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: rustDoc,
		Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
			// CC_RUST_FEATURES
			// https://doc.rust-lang.org/cargo/reference/features.html#command-line-feature-options
			// Multiple features may be separated with commas or spaces.
			// If using spaces, be sure to use quotes around all the features if running Cargo from a shell (such as --features "foo bar").
			// If building multiple packages in a workspace, the package-name/feature-name syntax can be used to specify features for specific workspace members.
			"features": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of Rust features to enable during build",
			},
			// CC_RUN_COMMAND
			"run_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom command to run your application",
			},
			// CC_RUST_BIN
			"rust_bin": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of the binary to launch once built",
			},
			// CC_RUSTUP_CHANNEL
			"rustup_channel": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The rust channel to use. Use a specific channel version with `stable`, `beta`, `nightly` or a specific version like `1.13.0` (default: `stable`)",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (r Rust) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := r.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	// Handle Rust features
	features := r.FeaturesAsStrings(ctx, diags)
	if len(features) > 0 {
		env[CC_RUST_FEATURES] = strings.Join(features, ",")
	}

	pkg.IfIsSetStr(r.RunCommand, func(s string) { env[CC_RUN_COMMAND] = s })
	pkg.IfIsSetStr(r.RustBin, func(s string) { env[CC_RUST_BIN] = s })
	pkg.IfIsSetStr(r.RustupChannel, func(s string) { env[CC_RUSTUP_CHANNEL] = s })

	return env
}

func (r *Rust) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	// Handle Rust features - convert comma-separated string to Set
	if featuresStr := m.Pop(CC_RUST_FEATURES); featuresStr != "" {
		features := strings.Split(featuresStr, ",")
		r.Features, _ = types.SetValueFrom(ctx, types.StringType, features)
	}

	r.RunCommand = pkg.FromStr(m.Pop(CC_RUN_COMMAND))
	r.RustBin = pkg.FromStr(m.Pop(CC_RUST_BIN))
	r.RustupChannel = pkg.FromStr(m.Pop(CC_RUSTUP_CHANNEL))

	r.FromEnvironment(ctx, m)
}

func (r Rust) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if r.Deployment == nil || r.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    r.Deployment.Repository.ValueString(),
		Commit:        r.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
