package dotnet

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

type Dotnet struct {
	attributes.Runtime
	DotnetProfile types.String `tfsdk:"dotnet_profile"`
	DotnetProj    types.String `tfsdk:"dotnet_proj"`
	DotnetTFM     types.String `tfsdk:"dotnet_tfm"`
	DotnetVersion types.String `tfsdk:"dotnet_version"`
}

//go:embed doc.md
var dotnetDoc string

func (r ResourceDotnet) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: dotnetDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			"dotnet_profile": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Override the build configuration settings in your project. Default: Release",
			},
			"dotnet_proj": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The name of your project file to use for the build, without the .csproj / .fsproj / .vbproj extension.",
			},
			"dotnet_tfm": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Compiles for a specific framework. The framework must be defined in the project file. Example : net5.0",
			},
			"dotnet_version": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Choose the .NET Core version between 6.0, 8.0, 9.0. Default: '8.0'",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceDotnet) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (dotnetapp Dotnet) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(dotnetapp.Environment.ElementsAs(ctx, &customEnv, false)...)

	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(dotnetapp.DotnetProfile, func(s string) { env["CC_DOTNET_PROFILE"] = s })
	pkg.IfIsSetStr(dotnetapp.DotnetProj, func(s string) { env["CC_DOTNET_PROJ"] = s })
	pkg.IfIsSetStr(dotnetapp.DotnetTFM, func(s string) { env["CC_DOTNET_TFM"] = s })
	pkg.IfIsSetStr(dotnetapp.DotnetVersion, func(s string) { env["CC_DOTNET_VERSION"] = s })

	env = pkg.Merge(env, dotnetapp.Hooks.ToEnv())

	return env
}

func (dotnetapp Dotnet) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if dotnetapp.Deployment == nil || dotnetapp.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    dotnetapp.Deployment.Repository.ValueString(),
		Commit:        dotnetapp.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
