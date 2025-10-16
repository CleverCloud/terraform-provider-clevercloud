package php

import (
	"context"
	_ "embed"
	"fmt"
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

type PHP struct {
	attributes.Runtime
	AlwaysPopulateRawPostData              types.String `tfsdk:"always_populate_raw_post_data"`
	ComposerVersion                        types.String `tfsdk:"composer_version"`
	CgiImplementation                      types.String `tfsdk:"cgi_implementation"`
	HttpBasicAuth                          types.String `tfsdk:"http_basic_auth"`
	ApacheHeadersSize                      types.Int64  `tfsdk:"apache_headers_size"`
	LdapCaCert                             types.String `tfsdk:"ldap_ca_cert"`
	MtaAuthPassword                        types.String `tfsdk:"mta_auth_password"`
	MtaAuthUser                            types.String `tfsdk:"mta_auth_user"`
	MtaServerAuthMethod                    types.String `tfsdk:"mta_server_auth_method"`
	MtaServerHost                          types.String `tfsdk:"mta_server_host"`
	MtaServerPort                          types.Int64  `tfsdk:"mta_server_port"`
	MtaServerUseTLS                        types.Bool   `tfsdk:"mta_server_use_tls"`
	OpcacheInternedStringsBuffer           types.Int64  `tfsdk:"opcache_interned_strings_buffer"`
	OpcacheMaxAcceleratedFiles             types.Int64  `tfsdk:"opcache_max_accelerated_files"`
	OpcacheMemory                          types.String `tfsdk:"opcache_memory"`
	OpcachePreload                         types.String `tfsdk:"opcache_preload"`
	AsyncAppBucket                         types.String `tfsdk:"async_app_bucket"`
	DevDependencies                        types.String `tfsdk:"dev_dependencies"`
	DisableAppBucket                       types.String `tfsdk:"disable_app_bucket"`
	PHPVersion                             types.String `tfsdk:"php_version"`
	RealpathCacheTTL                       types.Int64  `tfsdk:"realpath_cache_ttl"`
	WebRoot                                types.String `tfsdk:"webroot"`
	EnableElasticApmAgent                  types.Bool   `tfsdk:"enable_elastic_apm_agent"`
	EnableGrpc                             types.Bool   `tfsdk:"enable_grpc"`
	EnablePdflib                           types.Bool   `tfsdk:"enable_pdflib"`
	EnableRedis                            types.Bool   `tfsdk:"enable_redis"`
	HttpTimeout                            types.Int64  `tfsdk:"http_timeout"`
	LdaptlsCacert                          types.String `tfsdk:"ldaptls_cacert"`
	MaxInputVars                           types.Int64  `tfsdk:"max_input_vars"`
	MemoryLimit                            types.String `tfsdk:"memory_limit"`
	SessionType                            types.String `tfsdk:"session_type"`
	SocksifyEverything                     types.Bool   `tfsdk:"socksify_everything"`
	SqreenApiAppName                       types.String `tfsdk:"sqreen_api_app_name"`
	SqreenApiToken                         types.String `tfsdk:"sqreen_api_token"`
}

type PHPV0 struct {
	attributes.RuntimeV0
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
	Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
		"always_populate_raw_post_data": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Controls population of raw POST data",
		},
		"composer_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Choose your composer version between 1 and 2. Default is `2`",
		},
		"cgi_implementation": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Choose the Apache FastCGI module between `fastcgi` and `proxy_fcgi`. Default is `proxy_fcgi`",
		},
		"http_basic_auth": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Restrict HTTP access to your application. Example: `login:password`. You can define multiple credentials using additional `CC_HTTP_BASIC_AUTH_n` (where `n` is a number) environment variables",
		},
		"apache_headers_size": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Set the maximum size of the headers in Apache, between `8` and `256`. Default is `8`",
		},
		"ldap_ca_cert": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path to the LDAP CA certificate",
		},
		"mta_auth_password": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Password to authenticate to the SMTP server",
		},
		"mta_auth_user": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "User to authenticate to the SMTP server",
		},
		"mta_server_auth_method": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Enable or disable authentication to the SMTP server. Default is `on`",
		},
		"mta_server_host": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Host of the SMTP server",
		},
		"mta_server_port": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Port of the SMTP server. Default is `465`",
		},
		"mta_server_use_tls": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable or disable TLS when connecting to the SMTP server. Default is `true`",
		},
		"opcache_interned_strings_buffer": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The amount of memory used to store interned strings, in megabytes. Default is `4` (PHP5), `8` (PHP7)",
		},
		"opcache_max_accelerated_files": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum number of files handled by opcache. Default depends on the scaler size",
		},
		"opcache_memory": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Set the shared opcache memory size. Default is about 1/8 of the RAM",
		},
		"opcache_preload": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The path of the PHP preload file (PHP version 7.4 or higher)",
		},
		"async_app_bucket": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Mount the default app FS bucket asynchronously. If set, should have value `async`",
		},
		"dev_dependencies": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Control if development dependencies are installed or not. Values are either `install` or `ignore`",
		},
		"disable_app_bucket": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Disable entirely the app FS Bucket. Values are either `true`, `yes` or `disable`",
		},
		"php_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Choose your PHP version among those supported. Default is `8.3`",
		},
		"realpath_cache_ttl": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "The size of the realpath cache to be used by PHP. Default is `120`",
		},
		"webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Define the DocumentRoot of your project. Default is `.`",
		},
		"enable_elastic_apm_agent": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable the Elastic APM Agent for PHP. Default is `true` if `ELASTIC_APM_SERVER_URL` is defined, `false` otherwise",
		},
		"enable_grpc": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable the use of gRPC module. Default is `false`",
		},
		"enable_pdflib": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable the use of PDFlib module. Default is `false`",
		},
		"enable_redis": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable Redis support. Default is `false`",
		},
		"http_timeout": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Define a custom HTTP timeout. Default is `180`",
		},
		"ldaptls_cacert": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path to the LDAP TLS CA certificate",
		},
		"max_input_vars": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Maximum number of input variables that can be accepted",
		},
		"memory_limit": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Change the default memory limit for PHP scripts",
		},
		"session_type": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Choose `redis` to use Redis as session store",
		},
		"socksify_everything": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable SOCKS proxy for all outgoing connections. Default is `false`",
		},
		"sqreen_api_app_name": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The name of your Sqreen application",
		},
		"sqreen_api_token": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Your Sqreen organization token",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaPHPV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: phpDoc,
	Attributes: attributes.WithRuntimeCommonsV0(map[string]schema.Attribute{
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

func (p *PHP) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	// Start with common runtime environment variables (APP_FOLDER, Hooks, Environment)
	env := p.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	// Add PHP-specific environment variables
	pkg.IfIsSetStr(p.AlwaysPopulateRawPostData, func(s string) { env[ALWAYS_POPULATE_RAW_POST_DATA] = s })
	pkg.IfIsSetStr(p.ComposerVersion, func(s string) { env[CC_COMPOSER_VERSION] = s })
	pkg.IfIsSetStr(p.CgiImplementation, func(s string) { env[CC_CGI_IMPLEMENTATION] = s })
	pkg.IfIsSetStr(p.HttpBasicAuth, func(s string) { env[CC_HTTP_BASIC_AUTH] = s })
	pkg.IfIsSetI(p.ApacheHeadersSize, func(i int64) { env[CC_APACHE_HEADERS_SIZE] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(p.LdapCaCert, func(s string) { env[CC_LDAP_CA_CERT] = s })
	pkg.IfIsSetStr(p.MtaAuthPassword, func(s string) { env[CC_MTA_AUTH_PASSWORD] = s })
	pkg.IfIsSetStr(p.MtaAuthUser, func(s string) { env[CC_MTA_AUTH_USER] = s })
	pkg.IfIsSetStr(p.MtaServerAuthMethod, func(s string) { env[CC_MTA_SERVER_AUTH_METHOD] = s })
	pkg.IfIsSetStr(p.MtaServerHost, func(s string) { env[CC_MTA_SERVER_HOST] = s })
	pkg.IfIsSetI(p.MtaServerPort, func(i int64) { env[CC_MTA_SERVER_PORT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetB(p.MtaServerUseTLS, func(b bool) { env[CC_MTA_SERVER_USE_TLS] = strconv.FormatBool(b) })
	pkg.IfIsSetI(p.OpcacheInternedStringsBuffer, func(i int64) { env[CC_OPCACHE_INTERNED_STRINGS_BUFFER] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetI(p.OpcacheMaxAcceleratedFiles, func(i int64) { env[CC_OPCACHE_MAX_ACCELERATED_FILES] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(p.OpcacheMemory, func(s string) { env[CC_OPCACHE_MEMORY] = s })
	pkg.IfIsSetStr(p.OpcachePreload, func(s string) { env[CC_OPCACHE_PRELOAD] = s })
	pkg.IfIsSetStr(p.AsyncAppBucket, func(s string) { env[CC_PHP_ASYNC_APP_BUCKET] = s })
	pkg.IfIsSetStr(p.DevDependencies, func(s string) { env[CC_PHP_DEV_DEPENDENCIES] = s })
	pkg.IfIsSetStr(p.DisableAppBucket, func(s string) { env[CC_PHP_DISABLE_APP_BUCKET] = s })
	pkg.IfIsSetStr(p.PHPVersion, func(s string) { env[CC_PHP_VERSION] = s })
	pkg.IfIsSetI(p.RealpathCacheTTL, func(i int64) { env[CC_REALPATH_CACHE_TTL] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(p.WebRoot, func(s string) { env[CC_WEBROOT] = s })
	pkg.IfIsSetB(p.EnableElasticApmAgent, func(b bool) { env[ENABLE_ELASTIC_APM_AGENT] = strconv.FormatBool(b) })
	pkg.IfIsSetB(p.EnableGrpc, func(b bool) { env[ENABLE_GRPC] = strconv.FormatBool(b) })
	pkg.IfIsSetB(p.EnablePdflib, func(b bool) { env[ENABLE_PDFLIB] = strconv.FormatBool(b) })
	pkg.IfIsSetB(p.EnableRedis, func(b bool) { env[ENABLE_REDIS] = strconv.FormatBool(b) })
	pkg.IfIsSetI(p.HttpTimeout, func(i int64) { env[HTTP_TIMEOUT] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(p.LdaptlsCacert, func(s string) { env[LDAPTLS_CACERT] = s })
	pkg.IfIsSetI(p.MaxInputVars, func(i int64) { env[MAX_INPUT_VARS] = fmt.Sprintf("%d", i) })
	pkg.IfIsSetStr(p.MemoryLimit, func(s string) { env[MEMORY_LIMIT] = s })
	pkg.IfIsSetStr(p.SessionType, func(s string) { env[SESSION_TYPE] = s })
	pkg.IfIsSetB(p.SocksifyEverything, func(b bool) { env[SOCKSIFY_EVERYTHING] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(p.SqreenApiAppName, func(s string) { env[SQREEN_API_APP_NAME] = s })
	pkg.IfIsSetStr(p.SqreenApiToken, func(s string) { env[SQREEN_API_TOKEN] = s })

	return env
}

// fromEnv iter on environment set on the clever application and
// handle language specific env vars
// put the others on Environment field
func (p *PHP) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	// Parse PHP-specific environment variables
	p.AlwaysPopulateRawPostData = pkg.FromStr(m.Pop(ALWAYS_POPULATE_RAW_POST_DATA))
	p.ComposerVersion = pkg.FromStr(m.Pop(CC_COMPOSER_VERSION))
	p.CgiImplementation = pkg.FromStr(m.Pop(CC_CGI_IMPLEMENTATION))
	p.HttpBasicAuth = pkg.FromStr(m.Pop(CC_HTTP_BASIC_AUTH))

	if size, err := strconv.ParseInt(m.Pop(CC_APACHE_HEADERS_SIZE), 10, 64); err == nil {
		p.ApacheHeadersSize = pkg.FromI(size)
	}

	p.LdapCaCert = pkg.FromStr(m.Pop(CC_LDAP_CA_CERT))
	p.MtaAuthPassword = pkg.FromStr(m.Pop(CC_MTA_AUTH_PASSWORD))
	p.MtaAuthUser = pkg.FromStr(m.Pop(CC_MTA_AUTH_USER))
	p.MtaServerAuthMethod = pkg.FromStr(m.Pop(CC_MTA_SERVER_AUTH_METHOD))
	p.MtaServerHost = pkg.FromStr(m.Pop(CC_MTA_SERVER_HOST))

	if port, err := strconv.ParseInt(m.Pop(CC_MTA_SERVER_PORT), 10, 64); err == nil {
		p.MtaServerPort = pkg.FromI(port)
	}

	p.MtaServerUseTLS = pkg.FromBool(m.Pop(CC_MTA_SERVER_USE_TLS) == "true")

	if buffer, err := strconv.ParseInt(m.Pop(CC_OPCACHE_INTERNED_STRINGS_BUFFER), 10, 64); err == nil {
		p.OpcacheInternedStringsBuffer = pkg.FromI(buffer)
	}

	if files, err := strconv.ParseInt(m.Pop(CC_OPCACHE_MAX_ACCELERATED_FILES), 10, 64); err == nil {
		p.OpcacheMaxAcceleratedFiles = pkg.FromI(files)
	}

	p.OpcacheMemory = pkg.FromStr(m.Pop(CC_OPCACHE_MEMORY))
	p.OpcachePreload = pkg.FromStr(m.Pop(CC_OPCACHE_PRELOAD))
	p.AsyncAppBucket = pkg.FromStr(m.Pop(CC_PHP_ASYNC_APP_BUCKET))
	p.DevDependencies = pkg.FromStr(m.Pop(CC_PHP_DEV_DEPENDENCIES))
	p.DisableAppBucket = pkg.FromStr(m.Pop(CC_PHP_DISABLE_APP_BUCKET))
	p.PHPVersion = pkg.FromStr(m.Pop(CC_PHP_VERSION))

	if ttl, err := strconv.ParseInt(m.Pop(CC_REALPATH_CACHE_TTL), 10, 64); err == nil {
		p.RealpathCacheTTL = pkg.FromI(ttl)
	}

	p.WebRoot = pkg.FromStr(m.Pop(CC_WEBROOT))
	p.EnableElasticApmAgent = pkg.FromBool(m.Pop(ENABLE_ELASTIC_APM_AGENT) == "true")
	p.EnableGrpc = pkg.FromBool(m.Pop(ENABLE_GRPC) == "true")
	p.EnablePdflib = pkg.FromBool(m.Pop(ENABLE_PDFLIB) == "true")
	p.EnableRedis = pkg.FromBool(m.Pop(ENABLE_REDIS) == "true")

	if timeout, err := strconv.ParseInt(m.Pop(HTTP_TIMEOUT), 10, 64); err == nil {
		p.HttpTimeout = pkg.FromI(timeout)
	}

	p.LdaptlsCacert = pkg.FromStr(m.Pop(LDAPTLS_CACERT))

	if vars, err := strconv.ParseInt(m.Pop(MAX_INPUT_VARS), 10, 64); err == nil {
		p.MaxInputVars = pkg.FromI(vars)
	}

	p.MemoryLimit = pkg.FromStr(m.Pop(MEMORY_LIMIT))
	p.SessionType = pkg.FromStr(m.Pop(SESSION_TYPE))
	p.SocksifyEverything = pkg.FromBool(m.Pop(SOCKSIFY_EVERYTHING) == "true")
	p.SqreenApiAppName = pkg.FromStr(m.Pop(SQREEN_API_APP_NAME))
	p.SqreenApiToken = pkg.FromStr(m.Pop(SQREEN_API_TOKEN))

	// Handle common runtime variables (APP_FOLDER, Hooks, remaining Environment)
	p.FromEnvironment(ctx, m)
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
