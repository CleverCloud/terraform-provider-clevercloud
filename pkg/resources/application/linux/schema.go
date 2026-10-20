package linux

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

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
				MarkdownDescription: "The command to start your application (`CC_RUN_COMMAND`).",
			},
			"build_command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The command to run during the build phase (`CC_BUILD_COMMAND`).",
			},
			"makefile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom Makefile name or path (`CC_MAKEFILE`).",
			},
			"mise_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom path for the mise.toml configuration file (relative path) (`CC_MISE_FILE_PATH`).",
			},
			"disable_mise": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Disable Mise tool installation (Default: false) (`CC_DISABLE_MISE`).",
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

	pkg.WriteEnv(env, l, envVars...)

	env = pkg.Merge(env, l.Hooks.ToEnv())
	env = pkg.Merge(env, l.Integrations.ToEnv(ctx, diags))

	return env
}

func (l *Linux) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if l == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, l, envVars...)

	l.Integrations = attributes.FromEnvIntegrations(ctx, env, l.Integrations, diags)
}
