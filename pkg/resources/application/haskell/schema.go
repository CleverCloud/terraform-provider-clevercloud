package haskell

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
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Haskell struct {
	application.Runtime
	StackTarget                     types.String `tfsdk:"stack_target"`
	StackSetupCommand               types.String `tfsdk:"stack_setup_command"`
	StackInstallCommand             types.String `tfsdk:"stack_install_command"`
	StackInstallDependenciesCommand types.String `tfsdk:"stack_install_dependencies_command"`
}

//go:embed doc.md
var haskellDoc string

func (r ResourceHaskell) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: haskellDoc,
		Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
			"stack_target": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Specify Stack package target.",
			},
			"stack_setup_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only use this variable to override the default `setup` Stack step command.",
			},
			"stack_install_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only use this variable to override the default `install` Stack step command.",
			},
			"stack_install_dependencies_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only use this variable to override the default `install --only-dependencies` Stack step command.",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (haskellapp Haskell) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(haskellapp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(haskellapp.StackTarget, func(s string) { env["CC_HASKELL_STACK_TARGET"] = s })
	pkg.IfIsSetStr(haskellapp.StackSetupCommand, func(s string) { env["CC_HASKELL_STACK_SETUP_COMMAND"] = s })
	pkg.IfIsSetStr(haskellapp.StackInstallCommand, func(s string) { env["CC_HASKELL_STACK_INSTALL_COMMAND"] = s })
	pkg.IfIsSetStr(haskellapp.StackInstallDependenciesCommand, func(s string) { env["CC_HASKELL_STACK_INSTALL_DEPENDENCIES_COMMAND"] = s })

	env = pkg.Merge(env, haskellapp.Hooks.ToEnv())
	env = pkg.Merge(env, haskellapp.Integrations.ToEnv(ctx, diags))

	return env
}

func (haskellapp *Haskell) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	haskellapp.StackTarget = pkg.FromStrPtr(env.PopPtr("CC_HASKELL_STACK_TARGET"))
	haskellapp.StackSetupCommand = pkg.FromStrPtr(env.PopPtr("CC_HASKELL_STACK_SETUP_COMMAND"))
	haskellapp.StackInstallCommand = pkg.FromStrPtr(env.PopPtr("CC_HASKELL_STACK_INSTALL_COMMAND"))
	haskellapp.StackInstallDependenciesCommand = pkg.FromStrPtr(env.PopPtr("CC_HASKELL_STACK_INSTALL_DEPENDENCIES_COMMAND"))

	haskellapp.Integrations = attributes.FromEnvIntegrations(ctx, env, haskellapp.Integrations, diags)
}

func (haskellapp Haskell) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if haskellapp.Deployment == nil || haskellapp.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    haskellapp.Deployment.Repository.ValueString(),
		Commit:        haskellapp.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !haskellapp.Deployment.BasicAuthentication.IsNull() && !haskellapp.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema valisation step
		userPass := haskellapp.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
