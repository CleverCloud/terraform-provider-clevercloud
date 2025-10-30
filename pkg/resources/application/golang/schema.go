package golang

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

type Go struct {
	application.Runtime
	GoBuildTool types.String `tfsdk:"go_build_tool"`
	GoPkg       types.String `tfsdk:"go_pkg"`
	GoRunDir    types.String `tfsdk:"go_rundir"`
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
		// CC_GO_BUILD_TOOL
		"go_build_tool": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Available values: `gomod`, `gobuild`. Build and install your application (`goget` is deprecated)",
		},
		// CC_GO_PKG
		"go_pkg": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Tell the `CC_GO_BUILD_TOOL` which file contains the `main()` function (default: `main.go`)",
		},
		// CC_GO_RUNDIR
		"go_rundir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Run the application from the specified path, relative to `$GOPATH/src/` (deprecated)",
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

func (g Go) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := g.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	pkg.IfIsSetStr(g.GoBuildTool, func(s string) { env[CC_GO_BUILD_TOOL] = s })
	pkg.IfIsSetStr(g.GoPkg, func(s string) { env[CC_GO_PKG] = s })
	pkg.IfIsSetStr(g.GoRunDir, func(s string) { env[CC_GO_RUNDIR] = s })

	return env
}

func (g *Go) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	g.GoBuildTool = pkg.FromStr(m.Pop(CC_GO_BUILD_TOOL))
	g.GoPkg = pkg.FromStr(m.Pop(CC_GO_PKG))
	g.GoRunDir = pkg.FromStr(m.Pop(CC_GO_RUNDIR))

	g.FromEnvironment(ctx, m)
}

func (g Go) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if g.Deployment == nil || g.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    g.Deployment.Repository.ValueString(),
		Commit:        g.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
