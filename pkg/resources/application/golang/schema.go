package golang

import (
	"context"
	_ "embed"

	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Go struct {
	application.Runtime
	Binary    types.String `tfsdk:"binary"`
	BuildTool types.String `tfsdk:"build_tool"`
	Package   types.String `tfsdk:"package"`
	RunDir    types.String `tfsdk:"run_dir"`
}

type GoV0 struct {
	application.RuntimeV0
}

//go:embed doc.md
var goDoc string

func (r ResourceGo) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaGo
}

var schemaGo = schema.Schema{
	Version:             1,
	MarkdownDescription: goDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"binary": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path of the binary to run after a `makefile` build (`CC_GO_BINARY`)",
		},
		"build_tool": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "How to build the application: `gomod`, `gobuild`, `goget` or `makefile` (`CC_GO_BUILD_TOOL`)",
		},
		"package": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Import path of the Go package to build, detected from `go.mod` by default (`CC_GO_PKG`)",
		},
		"run_dir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Directory, within `GOPATH/src`, to place and run the application from (`CC_GO_RUNDIR`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaGoV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: goDoc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (g Go) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(g.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.WriteEnv(env, &g, envVars...)
	env = pkg.Merge(env, g.Hooks.ToEnv())
	env = pkg.Merge(env, g.Integrations.ToEnv(ctx, diags))

	return env
}

func (g *Go) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if g == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, g, envVars...)

	g.Integrations = attributes.FromEnvIntegrations(ctx, env, g.Integrations, diags)
}
