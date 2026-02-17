package linux

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Linux struct {
	application.Runtime
	RunCommand   types.String `tfsdk:"run_command"`
	BuildCommand types.String `tfsdk:"build_command"`
	Makefile     types.String `tfsdk:"makefile"`
	MiseFilePath types.String `tfsdk:"mise_file_path"`
	DisableMise  types.Bool   `tfsdk:"disable_mise"`
}

//go:embed doc.md
var linuxDoc string

func (r ResourceLinux) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: linuxDoc,
		Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
			"run_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The command to start your application.",
			},
			"build_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The command to run during the build phase.",
			},
			"makefile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom Makefile name or path.",
			},
			"mise_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom path for the mise.toml configuration file (relative path).",
			},
			"disable_mise": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Disable Mise tool installation (Default: false).",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (l *Linux) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(l.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetStr(l.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSetStr(l.RunCommand, func(s string) { env["CC_RUN_COMMAND"] = s })
	pkg.IfIsSetStr(l.BuildCommand, func(s string) { env["CC_BUILD_COMMAND"] = s })
	pkg.IfIsSetStr(l.Makefile, func(s string) { env["CC_MAKEFILE"] = s })
	pkg.IfIsSetStr(l.MiseFilePath, func(s string) { env["CC_MISE_FILE_PATH"] = s })
	pkg.IfIsSetB(l.DisableMise, func(disable bool) {
		if disable {
			env["CC_DISABLE_MISE"] = "true"
		}
	})

	env = pkg.Merge(env, l.Hooks.ToEnv())
	env = pkg.Merge(env, l.Integrations.ToEnv(ctx, diags))

	return env
}

func (l *Linux) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	l.AppFolder = pkg.FromStrPtr(env.PopPtr("APP_FOLDER"))
	l.RunCommand = pkg.FromStrPtr(env.PopPtr("CC_RUN_COMMAND"))
	l.BuildCommand = pkg.FromStrPtr(env.PopPtr("CC_BUILD_COMMAND"))
	l.Makefile = pkg.FromStrPtr(env.PopPtr("CC_MAKEFILE"))
	l.MiseFilePath = pkg.FromStrPtr(env.PopPtr("CC_MISE_FILE_PATH"))
	pkg.SetBoolIf(&l.DisableMise, env.PopPtr("CC_DISABLE_MISE"), "true")

	l.Integrations = attributes.FromEnvIntegrations(ctx, env, l.Integrations, diags)
}

func (l *Linux) ToDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if l.Deployment == nil || l.Deployment.Repository.IsNull() {
		return nil
	}

	d := &application.Deployment{
		Repository:    l.Deployment.Repository.ValueString(),
		Commit:        l.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}

	if !l.Deployment.BasicAuthentication.IsNull() && !l.Deployment.BasicAuthentication.IsUnknown() {
		// Expect validation to be done in the schema validation step
		userPass := l.Deployment.BasicAuthentication.ValueString()
		splits := strings.SplitN(userPass, ":", 2)
		d.Username = &splits[0]
		d.Password = &splits[1]
	}

	return d
}
