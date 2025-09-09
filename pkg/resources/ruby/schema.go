package ruby

import (
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type Ruby struct {
	attributes.Runtime
	RubyVersion types.String `tfsdk:"ruby_version"`
	BundlerFile types.String `tfsdk:"bundler_file"`
}

//go:embed doc.md
var rubyDoc string

func (r ResourceRuby) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: rubyDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			// CC_RUBY_VERSION
			"ruby_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Ruby version (Default: 3.1)",
			},
			// CC_BUNDLER_FILE
			"bundler_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Define a custom Gemfile path (default: Gemfile)",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceRuby) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (ruby Ruby) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(ruby.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSet(ruby.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSet(ruby.RubyVersion, func(version string) { env["CC_RUBY_VERSION"] = version })
	pkg.IfIsSet(ruby.BundlerFile, func(bundlerFile string) { env["CC_BUNDLER_FILE"] = bundlerFile })

	env = pkg.Merge(env, ruby.Hooks.ToEnv())
	return env
}

func (ruby Ruby) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if ruby.Deployment == nil || ruby.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    ruby.Deployment.Repository.ValueString(),
		Commit:        ruby.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
