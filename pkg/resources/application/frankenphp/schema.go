package frankenphp

import (
	"context"
	_ "embed"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type FrankenPHP struct {
	application.Runtime
	ListenedPort    types.Int64  `tfsdk:"listened_port"`
	WorkerPath      types.String `tfsdk:"worker_path"`
	ComposerFlags   types.String `tfsdk:"composer_flags"`
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
	Webroot         types.String `tfsdk:"webroot"`
}

//go:embed doc.md
var frankenphpDoc string

func (r ResourceFrankenPHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: frankenphpDoc,
		Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
			"listened_port": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The port on which FrankenPHP listens for HTTP requests",
			},
			"worker_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to the worker script, relative to the root of your project (e.g. /worker/scrip.php)",
			},
			"composer_flags": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Flags to pass to Composer",
			},
			"dev_dependencies": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Install development dependencies (Default: false)",
				Default:             booldefault.StaticBool(false),
			},
			"webroot": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to the web content to serve, relative to the root of your application",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

func (fp *FrankenPHP) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	// Start with common runtime environment variables (APP_FOLDER, Hooks, Environment)
	env := fp.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	// Add FrankenPHP-specific environment variables
	pkg.IfIsSetI(fp.ListenedPort, func(i int64) { env[CC_FRANKENPHP_PORT] = pkg.FromI(i).String() })
	pkg.IfIsSetStr(fp.WorkerPath, func(s string) { env[CC_FRANKENPHP_WORKER] = s })
	pkg.IfIsSetStr(fp.ComposerFlags, func(s string) { env[CC_PHP_COMPOSER_FLAGS] = s })
	pkg.IfIsSetStr(fp.Webroot, func(s string) { env[CC_WEBROOT] = s })

	pkg.IfIsSetB(fp.DevDependencies, func(devDeps bool) {
		if devDeps {
			env[CC_PHP_DEV_DEPENDENCIES] = "install"
		}
	})

	return env
}

// fromEnv iter on environment set on the clever application and
// handle language specific env vars
// put the others on Environment field
func (fp *FrankenPHP) fromEnv(ctx context.Context, env map[string]string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	m := helper.NewEnvMap(env)

	// Parse FrankenPHP-specific environment variables
	if port := m.Pop(CC_FRANKENPHP_PORT); port != "" {
		if parsed, err := strconv.ParseInt(port, 10, 64); err == nil {
			fp.ListenedPort = pkg.FromI(parsed)
		}
	}

	fp.WorkerPath = pkg.FromStr(m.Pop(CC_FRANKENPHP_WORKER))
	fp.ComposerFlags = pkg.FromStr(m.Pop(CC_PHP_COMPOSER_FLAGS))
	fp.Webroot = pkg.FromStr(m.Pop(CC_WEBROOT))

	if devDeps := m.Pop(CC_PHP_DEV_DEPENDENCIES); devDeps != "" {
		fp.DevDependencies = pkg.FromBool(devDeps == "install")
	}

	// Handle common runtime variables (APP_FOLDER, Hooks, remaining Environment)
	fp.FromEnvironment(ctx, m)
	return diags
}

func (fp *FrankenPHP) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if fp.Deployment == nil || fp.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    fp.Deployment.Repository.ValueString(),
		Commit:        fp.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
