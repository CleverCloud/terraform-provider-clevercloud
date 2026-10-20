package v

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

type V struct {
	application.Runtime
	Binary           types.String `tfsdk:"binary"`
	DevelopmentBuild types.Bool   `tfsdk:"development_build"`
	ModulesPath      types.String `tfsdk:"modules_path"`
}

//go:embed doc.md
var vDoc string

var schemaV = schema.Schema{
	Version:             1,
	MarkdownDescription: vDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"binary": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The name of the output binary file. Default: `${APP_HOME}/v_bin_${APP_ID}` (`CC_V_BINARY`)",
		},
		"development_build": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Set to true to compile without the `-prod` flag (`ENVIRONMENT`).",
		},
		"modules_path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Directory V looks up modules in, `/home/bas/.vmodules/` by default (`VMODULES`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (r ResourceV) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaV
}

func (vapp V) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(vapp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.WriteEnv(env, &vapp, envVars...)

	env = pkg.Merge(env, vapp.Hooks.ToEnv())
	env = pkg.Merge(env, vapp.Integrations.ToEnv(ctx, diags))

	return env
}

func (vapp *V) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if vapp == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, vapp, envVars...)

	vapp.Integrations = attributes.FromEnvIntegrations(ctx, env, vapp.Integrations, diags)
}
