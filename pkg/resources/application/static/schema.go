package static

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"

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

type Static struct {
	application.Runtime
	BuildCommand          types.String `tfsdk:"build_command"`
	HugoVersion           types.String `tfsdk:"hugo_version"`
	OverrideBuildcache    types.String `tfsdk:"override_buildcache"`
	StaticAutobuildOutDir types.String `tfsdk:"static_autobuild_outdir"`
	StaticCaddyfile       types.String `tfsdk:"static_caddyfile"`
	StaticFlags           types.String `tfsdk:"static_flags"`
	StaticPort            types.Int64  `tfsdk:"static_port"`
	StaticServer          types.String `tfsdk:"static_server"`
	WebRoot               types.String `tfsdk:"webroot"`
}

type StaticV0 struct {
	application.RuntimeV0
}

//go:embed doc.md
var staticDoc string

func (r ResourceStatic) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaStatic
}

var schemaStatic = schema.Schema{
	Version:             1,
	MarkdownDescription: staticDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		// CC_BUILD_COMMAND
		"build_command": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Command to run during build phase",
		},
		// CC_HUGO_VERSION
		"hugo_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Set Hugo version (e.g., `0.150`)",
		},
		// CC_OVERRIDE_BUILDCACHE
		"override_buildcache": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Customize build cache directories",
		},
		// CC_STATIC_AUTOBUILD_OUTDIR
		"static_autobuild_outdir": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Output directory for static site generator (default: `/cc_static_autobuilt`)",
		},
		// CC_STATIC_CADDYFILE
		"static_caddyfile": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path to Caddyfile for custom Caddy configuration (default: `./Caddyfile`)",
		},
		// CC_STATIC_FLAGS
		"static_flags": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Custom command line flags to pass to the static server",
		},
		// CC_STATIC_PORT
		"static_port": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Custom listen port for the static server (default: `8080`)",
		},
		// CC_STATIC_SERVER
		"static_server": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Server to use for static website (default: `static-web-server`)",
		},
		// CC_WEBROOT
		"webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path to web content to serve (default: `/`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaStaticV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: staticDoc,
	Attributes:          application.WithRuntimeCommonsV0(map[string]schema.Attribute{}),
	Blocks:              attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (plan *Static) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := plan.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	pkg.IfIsSetStr(plan.BuildCommand, func(s string) { env[CC_BUILD_COMMAND] = s })
	pkg.IfIsSetStr(plan.HugoVersion, func(s string) { env[CC_HUGO_VERSION] = s })
	pkg.IfIsSetStr(plan.OverrideBuildcache, func(s string) { env[CC_OVERRIDE_BUILDCACHE] = s })
	pkg.IfIsSetStr(plan.StaticAutobuildOutDir, func(s string) { env[CC_STATIC_AUTOBUILD_OUTDIR] = s })
	pkg.IfIsSetStr(plan.StaticCaddyfile, func(s string) { env[CC_STATIC_CADDYFILE] = s })
	pkg.IfIsSetStr(plan.StaticFlags, func(s string) { env[CC_STATIC_FLAGS] = s })
	pkg.IfIsSetI(plan.StaticPort, func(i int64) { env[CC_STATIC_PORT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(plan.StaticServer, func(s string) { env[CC_STATIC_SERVER] = s })
	pkg.IfIsSetStr(plan.WebRoot, func(s string) { env[CC_WEBROOT] = s })

	return env
}

func (static *Static) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	static.BuildCommand = pkg.FromStr(m.Pop(CC_BUILD_COMMAND))
	static.HugoVersion = pkg.FromStr(m.Pop(CC_HUGO_VERSION))
	static.OverrideBuildcache = pkg.FromStr(m.Pop(CC_OVERRIDE_BUILDCACHE))
	static.StaticAutobuildOutDir = pkg.FromStr(m.Pop(CC_STATIC_AUTOBUILD_OUTDIR))
	static.StaticCaddyfile = pkg.FromStr(m.Pop(CC_STATIC_CADDYFILE))
	static.StaticFlags = pkg.FromStr(m.Pop(CC_STATIC_FLAGS))

	if port, err := strconv.ParseInt(m.Pop(CC_STATIC_PORT), 10, 64); err == nil {
		static.StaticPort = pkg.FromI(port)
	}

	static.StaticServer = pkg.FromStr(m.Pop(CC_STATIC_SERVER))
	static.WebRoot = pkg.FromStr(m.Pop(CC_WEBROOT))

	static.FromEnvironment(ctx, m)
}

func (java *Static) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if java.Deployment == nil || java.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    java.Deployment.Repository.ValueString(),
		Commit:        java.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
