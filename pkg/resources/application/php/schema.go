package php

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

type PHP struct {
	application.Runtime
	PHPVersion      types.String `tfsdk:"php_version"`
	WebRoot         types.String `tfsdk:"webroot"`
	RedisSessions   types.Bool   `tfsdk:"redis_sessions"`
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`

	ComposerVersion              types.String `tfsdk:"composer_version"`
	CGIImplementation            types.String `tfsdk:"cgi_implementation"`
	ApacheHeadersSize            types.Int64  `tfsdk:"apache_headers_size"`
	HTTPTimeout                  types.Int64  `tfsdk:"http_timeout"`
	MaxInputVars                 types.Int64  `tfsdk:"max_input_vars"`
	MemoryLimit                  types.Int64  `tfsdk:"memory_limit"`
	RealpathCacheTTL             types.Int64  `tfsdk:"realpath_cache_ttl"`
	OpcacheMemory                types.Int64  `tfsdk:"opcache_memory"`
	OpcacheMaxAcceleratedFiles   types.Int64  `tfsdk:"opcache_max_accelerated_files"`
	OpcacheInternedStringsBuffer types.Int64  `tfsdk:"opcache_interned_strings_buffer"`
	OpcachePreload               types.String `tfsdk:"opcache_preload"`
	AsyncAppBucket               types.Bool   `tfsdk:"async_app_bucket"`
	DisableAppBucket             types.Bool   `tfsdk:"disable_app_bucket"`
	APMAgentEnabled              types.Bool   `tfsdk:"apm_agent_enabled"`
}

type PHPV0 struct {
	application.RuntimeV0
	PHPVersion      types.String `tfsdk:"php_version"`
	WebRoot         types.String `tfsdk:"webroot"`
	RedisSessions   types.Bool   `tfsdk:"redis_sessions"`
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
}

//go:embed doc.md
var phpDoc string

func (r ResourcePHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaPHP
}

var schemaPHP = schema.Schema{
	Version:             1,
	MarkdownDescription: phpDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		"php_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "PHP version (Default: 8) (`CC_PHP_VERSION`)",
		},
		"webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define the DocumentRoot of your project (default: \".\") (`CC_WEBROOT`)",
		},

		"redis_sessions": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Use a linked Redis instance to store sessions (Default: false) (`SESSION_TYPE`)",
		},
		"dev_dependencies": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Install development dependencies (`CC_PHP_DEV_DEPENDENCIES`)",
		},
		"composer_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Composer version installing the dependencies, `2` or `lts` (`CC_COMPOSER_VERSION`)",
		},
		"cgi_implementation": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Apache FastCGI module running PHP, `proxy_fcgi` (default) or `fastcgi` (`CC_CGI_IMPLEMENTATION`)",
		},
		"apache_headers_size": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum size of the HTTP request headers Apache accepts, in kilobytes, between 8 (default) and 256 (`CC_APACHE_HEADERS_SIZE`)",
		},
		"http_timeout": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Timeout of the HTTP requests between the reverse proxy and PHP, in seconds, 180 by default (`HTTP_TIMEOUT`)",
		},
		"max_input_vars": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum number of input variables PHP accepts per request (`MAX_INPUT_VARS`)",
		},
		"memory_limit": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum memory a PHP script can allocate, in MiB; computed from the flavor by default (`MEMORY_LIMIT`)",
		},
		"realpath_cache_ttl": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Time-to-live of the PHP realpath cache entries, in seconds, 120 by default (`CC_REALPATH_CACHE_TTL`)",
		},
		"opcache_memory": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Memory allocated to the OPcache, in MB; computed from the instance memory by default (`CC_OPCACHE_MEMORY`)",
		},
		"opcache_max_accelerated_files": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum number of PHP files stored in the OPcache; computed from the flavor by default (`CC_OPCACHE_MAX_ACCELERATED_FILES`)",
		},
		"opcache_interned_strings_buffer": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Memory used to store the OPcache interned strings, in MB (`CC_OPCACHE_INTERNED_STRINGS_BUFFER`)",
		},
		"opcache_preload": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path of a PHP script preloaded into the OPcache at server startup (`CC_OPCACHE_PRELOAD`)",
		},
		"async_app_bucket": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Mount the default application FS bucket asynchronously (`CC_PHP_ASYNC_APP_BUCKET`)",
		},
		"disable_app_bucket": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Do not mount the default application FS bucket at all (`CC_PHP_DISABLE_APP_BUCKET`)",
		},
		"apm_agent_enabled": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Enable the Elastic APM PHP extension; the platform enables it on its own when an Elastic APM server is configured, set `false` to prevent it (`ENABLE_ELASTIC_APM_AGENT`)",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaPHPV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: phpDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
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

func (p *PHP) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(p.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.WriteEnv(env, p, envVars...)
	env = pkg.Merge(env, p.Hooks.ToEnv())
	env = pkg.Merge(env, p.Integrations.ToEnv(ctx, diags))

	return env
}

func (p *PHP) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if p == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, p, envVars...)

	p.Integrations = attributes.FromEnvIntegrations(ctx, env, p.Integrations, diags)
}
