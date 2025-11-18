package rust

import (
	"context"
	_ "embed"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Rust struct {
	application.Runtime
	Features types.Set `tfsdk:"features"`
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
			// https://doc.rust-lang.org/cargo/reference/features.html#command-line-feature-options
			// Multiple features may be separated with commas or spaces.
			// If using spaces, be sure to use quotes around all the features if running Cargo from a shell (such as --features "foo bar").
			// If building multiple packages in a workspace, the package-name/feature-name syntax can be used to specify features for specific workspace members.
			"features": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of Rust features to enable during build",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (r Rust) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(r.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(r.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	env = pkg.Merge(env, r.Hooks.ToEnv())

	// Handle Rust features
	features := r.FeaturesAsStrings(ctx, diags)
	if len(features) > 0 {
		env[CC_RUST_FEATURES] = strings.Join(features, ",")
	}

	return env
}

func (r Rust) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if r.Deployment == nil || r.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    r.Deployment.Repository.ValueString(),
		Commit:        r.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !r.Deployment.BasicAuthentication.IsNull() && !r.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := r.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
