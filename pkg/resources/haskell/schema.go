package haskell

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

type Haskell struct {
	attributes.Runtime
	StackTarget types.String                     `tfsdk:"stack_target"`
	StackSetupCommand types.String               `tfsdk:"stack_setup_command"`
	StackInstallCommand types.String             `tfsdk:"stack_install_command"`
	StackInstallDependenciesCommand types.String `tfsdk:"stack_install_dependencies_command"`
}

//go:embed doc.md
var haskellDoc string

func (r ResourceHaskell) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: haskellDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			"stack_target": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Specify Stack package target.",
			},
			"stack_setup_command": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Only use this variable to override the default `setup` Stack step command.",
			},
			"stack_install_command": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Only use this variable to override the default `install` Stack step command.",
			},
			"stack_install_dependencies_command": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Only use this variable to override the default `install --only-dependencies` Stack step command",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceHaskell) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (haskellapp Haskell) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(haskellapp.Environment.ElementsAs(ctx, &customEnv, false)...)

	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(haskellapp.StackTarget, func(s string) { env["CC_HASKELL_STACK_TARGET"] = s })
	pkg.IfIsSetStr(haskellapp.StackSetupCommand, func(s string) { env["CC_HASKELL_STACK_SETUP_COMMAND"] = s })
	pkg.IfIsSetStr(haskellapp.StackInstallCommand, func(s string) { env["CC_HASKELL_STACK_INSTALL_COMMAND"] = s })
	pkg.IfIsSetStr(haskellapp.StackInstallDependenciesCommand, func(s string) { env["CC_HASKELL_STACK_INSTALL_DEPENDENCIES_COMMAND"] = s })

	env = pkg.Merge(env, haskellapp.Hooks.ToEnv())

	return env
}

func (haskellapp Haskell) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if haskellapp.Deployment == nil || haskellapp.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    haskellapp.Deployment.Repository.ValueString(),
		Commit:        haskellapp.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
