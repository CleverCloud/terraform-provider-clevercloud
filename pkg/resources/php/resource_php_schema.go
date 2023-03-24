package php

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

type PHP struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	MinInstanceCount types.Int64  `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64  `tfsdk:"max_instance_count"`
	SmallestFlavor   types.String `tfsdk:"smallest_flavor"`
	BiggestFlavor    types.String `tfsdk:"biggest_flavor"`
	BuildFlavor      types.String `tfsdk:"build_flavor"`
	Region           types.String `tfsdk:"region"`
	StickySessions   types.Bool   `tfsdk:"sticky_sessions"`
	RedirectHTTPS    types.Bool   `tfsdk:"redirect_https"`
	Commit           types.String `tfsdk:"commit"`
	VHost            types.String `tfsdk:"vhost"`
	AdditionalVHosts types.List   `tfsdk:"additional_vhosts"`
	DeployURL        types.String `tfsdk:"deploy_url"`

	// Env
	AppFolder   types.String `tfsdk:"app_folder"`
	Environment types.Map    `tfsdk:"environment"`

	// PHP related
	PHPVersion      types.String `tfsdk:"php_version"`
	WebRoot         types.String `tfsdk:"webroot"`
	RedisSessions   types.Bool   `tfsdk:"redis_sessions"`
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
}

//go:embed resource_php.md
var phpDoc string

func (r ResourcePHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: phpDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
			// CC_WEBROOT
			"php_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "PHP version (Default: 7)",
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
	for k, v := range customEnv {
		env[k] = v
	}

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

	return env
}
