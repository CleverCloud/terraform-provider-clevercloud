package dotnet

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

type Dotnet struct {
	application.Runtime
	DotnetProfile types.String `tfsdk:"profile"`
	DotnetProj    types.String `tfsdk:"proj"`
	DotnetTFM     types.String `tfsdk:"tfm"`
	DotnetVersion types.String `tfsdk:"version"`
}

//go:embed doc.md
var dotnetDoc string

func (r ResourceDotnet) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: dotnetDoc,
		Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
			"profile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override the build configuration settings in your project. Default: Release",
			},
			"proj": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name of your project file to use for the build, without the .csproj / .fsproj / .vbproj extension.",
			},
			"tfm": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Compiles for a specific framework. The framework must be defined in the project file. Example : net5.0",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Choose the .NET Core version between 6.0, 8.0, 9.0. Default: '8.0'",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (dotnetapp Dotnet) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
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

func (dotnetapp Dotnet) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if dotnetapp.Deployment == nil || dotnetapp.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    dotnetapp.Deployment.Repository.ValueString(),
		Commit:        dotnetapp.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !dotnetapp.Deployment.BasicAuthentication.IsNull() && !dotnetapp.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := dotnetapp.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
