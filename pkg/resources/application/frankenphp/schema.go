package frankenphp

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

type FrankenPHP struct {
	application.Runtime
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
	PHPVersion      types.String `tfsdk:"php_version"`
	WebRoot         types.String `tfsdk:"webroot"`
	ComposerVersion types.String `tfsdk:"composer_version"`
	ComposerFlags   types.String `tfsdk:"composer_flags"`
	Port            types.Int64  `tfsdk:"port"`
	WorkerScript    types.String `tfsdk:"worker_script"`
}

// FrankenPHPV0 is the schema version 0 shape of FrankenPHP, kept for state upgrades
type FrankenPHPV0 struct {
	application.RuntimeV0
	DevDependencies types.Bool `tfsdk:"dev_dependencies"`
}

//go:embed doc.md
var frankenphpDoc string

// schemaFrankenPHPV0 is the schema version 0 as shipped in 1.1.0, before vhosts became
// {fqdn, path_begin} objects. Frozen: it must keep decoding states written back then.
var schemaFrankenPHPV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: frankenphpDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
		"dev_dependencies": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Install development dependencies (Default: false)",
			Default:             booldefault.StaticBool(false),
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaFrankenPHP = schema.Schema{
	Version:             1,
	MarkdownDescription: frankenphpDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"dev_dependencies": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Install development dependencies (Default: false) (`CC_PHP_DEV_DEPENDENCIES`)",
			Default:             booldefault.StaticBool(false),
		},
		"php_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "PHP version to build and run the application with (`CC_PHP_VERSION`)",
		},
		"webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path to the web content to serve, relative to the application root (`CC_WEBROOT`)",
		},
		"composer_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Composer version installing the dependencies, `2` or `lts` (`CC_COMPOSER_VERSION`)",
		},
		"composer_flags": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Flags passed to `composer install`, replacing `-n --no-dev --no-progress --no-scripts` (`CC_PHP_COMPOSER_FLAGS`)",
		},
		"port": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Port FrankenPHP listens on for HTTP requests, 8080 by default; useful to run a service in front of FrankenPHP (`CC_FRANKENPHP_PORT`)",
		},
		"worker_script": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path of the worker script, relative to the application root, for example `/public/worker.php`; it runs FrankenPHP in worker mode (`CC_FRANKENPHP_WORKER`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (r ResourceFrankenPHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaFrankenPHP
}

func (fp *FrankenPHP) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(fp.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.WriteEnv(env, fp, envVars...)
	env = pkg.Merge(env, fp.Hooks.ToEnv())
	env = pkg.Merge(env, fp.Integrations.ToEnv(ctx, diags))

	return env
}

func (fp *FrankenPHP) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if fp == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, fp, envVars...)

	fp.Integrations = attributes.FromEnvIntegrations(ctx, env, fp.Integrations, diags)
}
