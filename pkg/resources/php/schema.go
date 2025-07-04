package php

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

type PHP struct {
	attributes.Runtime
	PHPVersion      types.String `tfsdk:"php_version"`
	WebRoot         types.String `tfsdk:"webroot"`
	RedisSessions   types.Bool   `tfsdk:"redis_sessions"`
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
}

//go:embed doc.md
var phpDoc string

func (r ResourcePHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: phpDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			// CC_WEBROOT
			"php_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "PHP version (Default: 8)",
			},
			"webroot": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Define the DocumentRoot of your project (default: \".\")",
			},

			"redis_sessions": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Use a linked Redis instance to store sessions (Default: false)",
			},
			"dev_dependencies": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Install development dependencies",
			},
		}),
		Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (p *PHP) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

func (p *PHP) toEnv(ctx context.Context, diags diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(p.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.IfIsSet(p.AppFolder, func(s string) { env["APP_FOLDER"] = s })
	pkg.IfIsSet(p.WebRoot, func(webroot string) { env["CC_WEBROOT"] = webroot })
	pkg.IfIsSet(p.PHPVersion, func(version string) { env["CC_PHP_VERSION"] = version })
	pkg.IfIsSetB(p.DevDependencies, func(devDeps bool) {
		if devDeps {
			env["CC_PHP_DEV_DEPENDENCIES"] = "install"
		}
	})
	pkg.IfIsSetB(p.RedisSessions, func(redis bool) {
		if redis {
			env["SESSION_TYPE"] = "redis"
		}
	})
	env = pkg.Merge(env, p.Hooks.ToEnv())

	return env
}

func (p *PHP) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if p.Deployment == nil || p.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    p.Deployment.Repository.ValueString(),
		Commit:        p.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
