package nodejs

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type NodeJS struct {
	application.Runtime
	DevDependencies      types.Bool   `tfsdk:"dev_dependencies"`
	StartScript          types.String `tfsdk:"start_script"`
	PackageManager       types.String `tfsdk:"package_manager"`
	Registry             types.String `tfsdk:"registry"`
	RegistryToken        types.String `tfsdk:"registry_token"`
	NodeVersion          types.String `tfsdk:"node_version"`
	CustomInstallCommand types.String `tfsdk:"custom_install_command"`
	RegistryBasicAuth    types.String `tfsdk:"registry_basic_auth"`
	BunInstallCacheDir   types.String `tfsdk:"bun_install_cache_dir"`
}

// envAttributes maps the environment variables backing a nodejs attribute to
// that attribute; ToEnv writes them and FromEnv reads them back.
var envAttributes = map[string]string{
	app.CC_NODE_DEV_DEPENDENCIES: "dev_dependencies",
	app.CC_RUN_COMMAND:           "start_script",
	app.CC_NODE_BUILD_TOOL:       "package_manager",
	app.CC_NPM_REGISTRY:          "registry",
	app.NPM_TOKEN:                "registry_token",
	app.CC_NODE_VERSION:          "node_version",
	app.CC_CUSTOM_BUILD_TOOL:     "custom_install_command",
	app.CC_NPM_BASIC_AUTH:        "registry_basic_auth",
	app.CC_BUN_INSTALL_CACHE_DIR: "bun_install_cache_dir",
}

type NodeJSV0 struct {
	application.RuntimeV0
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
	StartScript     types.String `tfsdk:"start_script"`
	PackageManager  types.String `tfsdk:"package_manager"`
	Registry        types.String `tfsdk:"registry"`
	RegistryToken   types.String `tfsdk:"registry_token"`
}

//go:embed doc.md
var nodejsDoc string

func (r ResourceNodeJS) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaNodeJS
}

var schemaNodeJS = schema.Schema{
	Version:             1,
	MarkdownDescription: nodejsDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
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
		// CC_NODE_BUILD_TOOL
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
			Sensitive:           !pkg.BypassSensitiveImports(),
			MarkdownDescription: "Private repository token",
		},
		// CC_NODE_VERSION
		"node_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Node.js version to use, for example `22`",
		},
		// CC_CUSTOM_BUILD_TOOL
		"custom_install_command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Command installing the dependencies, used when `package_manager` is `custom`",
		},
		// CC_NPM_BASIC_AUTH
		"registry_basic_auth": schema.StringAttribute{
			Optional:            true,
			Sensitive:           !pkg.BypassSensitiveImports(),
			MarkdownDescription: "Private repository credentials, as `user:password`. Cannot be used with `registry_token`",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("registry_token")),
			},
		},
		// CC_BUN_INSTALL_CACHE_DIR
		"bun_install_cache_dir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Bun install cache directory, saved and restored with the dependencies cache",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaNodeJSV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: nodejsDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
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
		// CC_NODE_BUILD_TOOL
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
			Sensitive:           !pkg.BypassSensitiveImports(),
			MarkdownDescription: "Private repository token",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (node NodeJS) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(node.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(node.AppFolder, func(s string) { env[app.APP_FOLDER] = s })
	pkg.IfIsSetB(node.DevDependencies, func(s bool) { env[app.CC_NODE_DEV_DEPENDENCIES] = "install" })
	pkg.IfIsSetStr(node.StartScript, func(s string) { env[app.CC_RUN_COMMAND] = s })
	pkg.IfIsSetStr(node.PackageManager, func(s string) { env[app.CC_NODE_BUILD_TOOL] = s })
	pkg.IfIsSetStr(node.Registry, func(s string) { env[app.CC_NPM_REGISTRY] = s })
	pkg.IfIsSetStr(node.RegistryToken, func(s string) { env[app.NPM_TOKEN] = s })
	pkg.IfIsSetStr(node.NodeVersion, func(s string) { env[app.CC_NODE_VERSION] = s })
	pkg.IfIsSetStr(node.CustomInstallCommand, func(s string) { env[app.CC_CUSTOM_BUILD_TOOL] = s })
	pkg.IfIsSetStr(node.RegistryBasicAuth, func(s string) { env[app.CC_NPM_BASIC_AUTH] = s })
	pkg.IfIsSetStr(node.BunInstallCacheDir, func(s string) { env[app.CC_BUN_INSTALL_CACHE_DIR] = s })
	env = pkg.Merge(env, node.Hooks.ToEnv())
	env = pkg.Merge(env, node.Integrations.ToEnv(ctx, diags))

	return env
}

func (node *NodeJS) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if node == nil || env == nil {
		return
	}

	node.AppFolder = pkg.FromStrPtr(env.PopPtr(app.APP_FOLDER))
	pkg.SetBoolIf(&node.DevDependencies, env.PopPtr(app.CC_NODE_DEV_DEPENDENCIES), "install")
	node.StartScript = pkg.FromStrPtr(env.PopPtr(app.CC_RUN_COMMAND))
	node.PackageManager = pkg.FromStrPtr(env.PopPtr(app.CC_NODE_BUILD_TOOL))
	node.Registry = pkg.FromStrPtr(env.PopPtr(app.CC_NPM_REGISTRY))
	node.RegistryToken = pkg.FromStrPtr(env.PopPtr(app.NPM_TOKEN))
	node.NodeVersion = pkg.FromStrPtr(env.PopPtr(app.CC_NODE_VERSION))
	node.CustomInstallCommand = pkg.FromStrPtr(env.PopPtr(app.CC_CUSTOM_BUILD_TOOL))
	node.RegistryBasicAuth = pkg.FromStrPtr(env.PopPtr(app.CC_NPM_BASIC_AUTH))
	node.BunInstallCacheDir = pkg.FromStrPtr(env.PopPtr(app.CC_BUN_INSTALL_CACHE_DIR))

	node.Integrations = attributes.FromEnvIntegrations(ctx, env, node.Integrations, diags)
}
