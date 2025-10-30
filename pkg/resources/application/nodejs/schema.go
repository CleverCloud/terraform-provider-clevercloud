package nodejs

import (
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type NodeJS struct {
	application.Runtime
	NodeVersion      types.String `tfsdk:"node_version"`
	DevDependencies  types.Bool   `tfsdk:"dev_dependencies"`
	StartScript      types.String `tfsdk:"start_script"`
	PackageManager   types.String `tfsdk:"package_manager"`
	CustomBuildTool  types.String `tfsdk:"custom_build_tool"`
	Registry         types.String `tfsdk:"registry"`
	RegistryBasicAuth types.String `tfsdk:"registry_basic_auth"`
	RegistryToken    types.String `tfsdk:"registry_token"`
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
		// CC_NODE_VERSION
		"node_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Set Node.js version, for example `24`, `23.11` or `22.15.1`",
		},
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
			MarkdownDescription: "Choose your build tool between npm, npm-ci, yarn, yarn2 and custom. Default is `npm`",
		},
		// CC_CUSTOM_BUILD_TOOL
		"custom_build_tool": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "A custom command to run (with package_manager set to `custom`)",
		},
		// CC_NPM_REGISTRY
		"registry": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The host of your private repository, available values: github or the registry host. Default is `registry.npmjs.org`",
		},
		// CC_NPM_BASIC_AUTH
		"registry_basic_auth": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Private repository credentials, in the form `user:password`. You can't use this if registry_token is set",
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

func (node NodeJS) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	// Start with common runtime environment variables (APP_FOLDER, Hooks, Environment)
	env := node.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	// Add Node.js-specific environment variables
	pkg.IfIsSetStr(node.NodeVersion, func(s string) { env[CC_NODE_VERSION] = s })
	pkg.IfIsSetB(node.DevDependencies, func(b bool) {
		if b {
			env[CC_NODE_DEV_DEPENDENCIES] = "install"
		}
	})
	pkg.IfIsSetStr(node.StartScript, func(s string) { env[CC_RUN_COMMAND] = s })
	pkg.IfIsSetStr(node.PackageManager, func(s string) { env[CC_NODE_BUILD_TOOL] = s })
	pkg.IfIsSetStr(node.CustomBuildTool, func(s string) { env[CC_CUSTOM_BUILD_TOOL] = s })
	pkg.IfIsSetStr(node.Registry, func(s string) { env[CC_NPM_REGISTRY] = s })
	pkg.IfIsSetStr(node.RegistryBasicAuth, func(s string) { env[CC_NPM_BASIC_AUTH] = s })
	pkg.IfIsSetStr(node.RegistryToken, func(s string) { env[NPM_TOKEN] = s })

	return env
}

// fromEnv iter on environment set on the clever application and
// handle language specific env vars
// put the others on Environment field
func (node *NodeJS) fromEnv(ctx context.Context, env map[string]string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	m := helper.NewEnvMap(env)

	// Parse Node.js-specific environment variables
	node.NodeVersion = pkg.FromStr(m.Pop(CC_NODE_VERSION))

	if devDeps := m.Pop(CC_NODE_DEV_DEPENDENCIES); devDeps != "" {
		node.DevDependencies = pkg.FromBool(devDeps == "install")
	}

	node.StartScript = pkg.FromStr(m.Pop(CC_RUN_COMMAND))
	node.PackageManager = pkg.FromStr(m.Pop(CC_NODE_BUILD_TOOL))
	node.CustomBuildTool = pkg.FromStr(m.Pop(CC_CUSTOM_BUILD_TOOL))
	node.Registry = pkg.FromStr(m.Pop(CC_NPM_REGISTRY))
	node.RegistryBasicAuth = pkg.FromStr(m.Pop(CC_NPM_BASIC_AUTH))
	node.RegistryToken = pkg.FromStr(m.Pop(NPM_TOKEN))

	// Handle common runtime variables (APP_FOLDER, Hooks, remaining Environment)
	node.FromEnvironment(ctx, m)
	return diags
}

func (node NodeJS) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if node.Deployment == nil || node.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    node.Deployment.Repository.ValueString(),
		Commit:        node.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
