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
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
	StartScript     types.String `tfsdk:"start_script"`
	PackageManager  types.String `tfsdk:"package_manager"`
	Registry        types.String `tfsdk:"registry"`
	RegistryToken   types.String `tfsdk:"registry_token"`
}

//go:embed doc.md
var dotnetDoc string

func (r ResourceDotnet) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: dotnetDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			// CC_NODE_DEV_DEPENDENCIES
			"dev_dependencies": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Install development dependencies specified in package.json",
			},
			// CC_RUN_COMMAND
			"start_script": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Set custom start script, instead of `npm start`",
			},
			// CC_NODE_BUILD_TOOL / CC_CUSTOM_BUILD_TOOL
			"package_manager": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Either npm, npm-ci, bun, pnpm, yarn-berry or custom",
			},
			// CC_NPM_REGISTRY
			"registry": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The host of your private repository, available values: github or the registry host",
			},
			// NPM_TOKEN
			"registry_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Private repository token",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceDotnet) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (dotnetapp Dotnet) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(dotnetapp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(dotnetapp.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSetB(dotnetapp.DevDependencies, func(s bool) { env["CC_NODE_DEV_DEPENDENCIES"] = "install" })
	pkg.IfIsSetStr(dotnetapp.StartScript, func(s string) { env["CC_RUN_COMMAND"] = s })
	pkg.IfIsSetStr(dotnetapp.PackageManager, func(s string) { env["CC_NODE_BUILD_TOOL"] = s })
	pkg.IfIsSetStr(dotnetapp.Registry, func(s string) { env["CC_NPM_REGISTRY"] = s })
	pkg.IfIsSetStr(dotnetapp.RegistryToken, func(s string) { env["NPM_TOKEN"] = s })
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
