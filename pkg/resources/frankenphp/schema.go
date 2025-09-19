package frankenphp

import (
	"context"
	_ "embed"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type FrankenPHP struct {
	attributes.Runtime
	DevDependencies types.Bool `tfsdk:"dev_dependencies"`
}

//go:embed doc.md
var frankenphpDoc string

func (r ResourceFrankenPHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: frankenphpDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			"dev_dependencies": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Install development dependencies (Default: false)",
				Default:             booldefault.StaticBool(false),
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourceFrankenPHP) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (fp *FrankenPHP) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(fp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSetB(fp.DevDependencies, func(devDeps bool) {
		if devDeps {
			env["CC_PHP_DEV_DEPENDENCIES"] = "install"
		}
	})
	env = pkg.Merge(env, fp.Hooks.ToEnv())

	return env
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
