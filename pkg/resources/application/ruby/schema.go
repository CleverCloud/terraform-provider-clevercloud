package ruby

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

type Ruby struct {
	application.Runtime
	RubyVersion           types.String `tfsdk:"ruby_version"`
	EnableSidekiq         types.Bool   `tfsdk:"enable_sidekiq"`
	RackupServer          types.String `tfsdk:"rackup_server"`
	RakeGoals             types.String `tfsdk:"rake_goals"`
	SidekiqFiles          types.String `tfsdk:"sidekiq_files"`
	HTTPBasicAuth         types.String `tfsdk:"http_basic_auth"`
	NginxProxyBuffers     types.String `tfsdk:"nginx_proxy_buffers"`
	NginxProxyBufferSize  types.String `tfsdk:"nginx_proxy_buffer_size"`
	EnableGzipCompression types.Bool   `tfsdk:"enable_gzip_compression"`
	GzipTypes             types.String `tfsdk:"gzip_types"`
	NginxReadTimeout      types.Int64  `tfsdk:"nginx_read_timeout"`
	RackEnv               types.String `tfsdk:"rack_env"`
	RailsEnv              types.String `tfsdk:"rails_env"`
	StaticFilesPath       types.String `tfsdk:"static_files_path"`
	StaticURLPrefix       types.String `tfsdk:"static_url_prefix"`
	StaticWebroot         types.String `tfsdk:"static_webroot"`
}

type RubyV0 struct {
	application.RuntimeV0
	RubyVersion           types.String `tfsdk:"ruby_version"`
	EnableSidekiq         types.Bool   `tfsdk:"enable_sidekiq"`
	RackupServer          types.String `tfsdk:"rackup_server"`
	RakeGoals             types.String `tfsdk:"rake_goals"`
	SidekiqFiles          types.String `tfsdk:"sidekiq_files"`
	HTTPBasicAuth         types.String `tfsdk:"http_basic_auth"`
	NginxProxyBuffers     types.String `tfsdk:"nginx_proxy_buffers"`
	NginxProxyBufferSize  types.String `tfsdk:"nginx_proxy_buffer_size"`
	EnableGzipCompression types.Bool   `tfsdk:"enable_gzip_compression"`
	GzipTypes             types.String `tfsdk:"gzip_types"`
	NginxReadTimeout      types.Int64  `tfsdk:"nginx_read_timeout"`
	RackEnv               types.String `tfsdk:"rack_env"`
	RailsEnv              types.String `tfsdk:"rails_env"`
	StaticFilesPath       types.String `tfsdk:"static_files_path"`
	StaticURLPrefix       types.String `tfsdk:"static_url_prefix"`
	StaticWebroot         types.String `tfsdk:"static_webroot"`
}

//go:embed doc.md
var rubyDoc string

func (r ResourceRuby) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schemaRuby
}

var schemaRuby = schema.Schema{
	Version:             1,
	MarkdownDescription: rubyDoc,
	Attributes: application.WithRuntimeCommons(map[string]schema.Attribute{
		// CC_RUBY_VERSION
		"ruby_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Ruby version to use (e.g., '3.3', '3.3.1') (`CC_RUBY_VERSION`)",
		},
		// CC_ENABLE_SIDEKIQ
		"enable_sidekiq": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Enable Sidekiq background process (`CC_ENABLE_SIDEKIQ`)",
		},
		// CC_RACKUP_SERVER
		"rackup_server": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Server to use for serving the Ruby application (default: puma) (`CC_RACKUP_SERVER`)",
		},
		// CC_RAKEGOALS
		"rake_goals": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Comma-separated list of rake goals to execute (e.g., 'db:migrate,assets:precompile') (`CC_RAKEGOALS`)",
		},
		// CC_SIDEKIQ_FILES
		"sidekiq_files": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Specify a list of Sidekiq configuration files (e.g., './config/sidekiq_1.yml,./config/sidekiq_2.yml') (`CC_SIDEKIQ_FILES`)",
		},
		// CC_HTTP_BASIC_AUTH
		"http_basic_auth": schema.StringAttribute{
			Optional:            true,
			Sensitive:           !pkg.BypassSensitiveImports(),
			MarkdownDescription: "Restrict HTTP access to your application (format: 'login:password') (`CC_HTTP_BASIC_AUTH`)",
		},
		// CC_NGINX_PROXY_BUFFERS
		"nginx_proxy_buffers": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Sets the number and size of the buffers used for reading a response from the proxied server (`CC_NGINX_PROXY_BUFFERS`)",
		},
		// CC_NGINX_PROXY_BUFFER_SIZE
		"nginx_proxy_buffer_size": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Sets the size of the buffer used for reading the first part of the response received from the proxied server (`CC_NGINX_PROXY_BUFFER_SIZE`)",
		},
		// ENABLE_GZIP_COMPRESSION
		"enable_gzip_compression": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Set to true to gzip-compress through Nginx (`ENABLE_GZIP_COMPRESSION`)",
		},
		// GZIP_TYPES
		"gzip_types": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Set the mime types to compress (default: 'text/* application/json application/xml application/javascript image/svg+xml') (`GZIP_TYPES`)",
		},
		// NGINX_READ_TIMEOUT
		"nginx_read_timeout": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Read timeout in seconds (default: 300) (`NGINX_READ_TIMEOUT`)",
		},
		// RACK_ENV
		"rack_env": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Rack environment variable (`RACK_ENV`)",
		},
		// RAILS_ENV
		"rails_env": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Rails environment variable (`RAILS_ENV`)",
		},
		// STATIC_FILES_PATH
		"static_files_path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Relative path to where your static files are stored (`STATIC_FILES_PATH`)",
		},
		// STATIC_URL_PREFIX
		"static_url_prefix": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL path under which you want to serve static files, usually /public (`STATIC_URL_PREFIX`)",
		},
		// STATIC_WEBROOT
		"static_webroot": schema.StringAttribute{
			Optional:            true,
			DeprecationMessage:  "has no effect, the platform does not read STATIC_WEBROOT: use `static_files_path` and `static_url_prefix`. It will be removed in a future release",
			MarkdownDescription: "**Deprecated**, has no effect: the platform does not read `STATIC_WEBROOT`. Use `static_files_path` and `static_url_prefix` instead, it will be removed in a future release.",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

var schemaRubyV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: rubyDoc,
	Attributes: application.WithRuntimeCommonsV0(map[string]schema.Attribute{
		// CC_RUBY_VERSION
		"ruby_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Ruby version to use (e.g., '3.3', '3.3.1')",
		},
		// CC_ENABLE_SIDEKIQ
		"enable_sidekiq": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Enable Sidekiq background process",
		},
		// CC_RACKUP_SERVER
		"rackup_server": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Server to use for serving the Ruby application (default: puma)",
		},
		// CC_RAKEGOALS
		"rake_goals": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Comma-separated list of rake goals to execute (e.g., 'db:migrate,assets:precompile')",
		},
		// CC_SIDEKIQ_FILES
		"sidekiq_files": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Specify a list of Sidekiq configuration files (e.g., './config/sidekiq_1.yml,./config/sidekiq_2.yml')",
		},
		// CC_HTTP_BASIC_AUTH
		"http_basic_auth": schema.StringAttribute{
			Optional:            true,
			Sensitive:           !pkg.BypassSensitiveImports(),
			MarkdownDescription: "Restrict HTTP access to your application (format: 'login:password')",
		},
		// CC_NGINX_PROXY_BUFFERS
		"nginx_proxy_buffers": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Sets the number and size of the buffers used for reading a response from the proxied server",
		},
		// CC_NGINX_PROXY_BUFFER_SIZE
		"nginx_proxy_buffer_size": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Sets the size of the buffer used for reading the first part of the response received from the proxied server",
		},
		// ENABLE_GZIP_COMPRESSION
		"enable_gzip_compression": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Set to true to gzip-compress through Nginx",
		},
		// GZIP_TYPES
		"gzip_types": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Set the mime types to compress (default: 'text/* application/json application/xml application/javascript image/svg+xml')",
		},
		// NGINX_READ_TIMEOUT
		"nginx_read_timeout": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Read timeout in seconds (default: 300)",
		},
		// RACK_ENV
		"rack_env": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Rack environment variable",
		},
		// RAILS_ENV
		"rails_env": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Rails environment variable",
		},
		// STATIC_FILES_PATH
		"static_files_path": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Relative path to where your static files are stored",
		},
		// STATIC_URL_PREFIX
		"static_url_prefix": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "The URL path under which you want to serve static files, usually /public",
		},
		// STATIC_WEBROOT
		"static_webroot": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Path to the web content to serve, relative to the root of your application",
		},
	}),
	Blocks: attributes.WithBlockRuntimeCommons(map[string]schema.Block{}),
}

func (ruby Ruby) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := map[string]string{}

	// do not use the real map since ElementAs can nullish it
	// https://github.com/hashicorp/terraform-plugin-framework/issues/698
	customEnv := map[string]string{}
	diags.Append(ruby.Environment.ElementsAs(ctx, &customEnv, false)...)
	if diags.HasError() {
		return env
	}
	env = pkg.Merge(env, customEnv)

	pkg.WriteEnv(env, &ruby, envVars...)
	// deprecated, the platform ignores it: kept out of envVars
	pkg.IfIsSetStr(ruby.StaticWebroot, func(s string) { env["STATIC_WEBROOT"] = s })
	env = pkg.Merge(env, ruby.Hooks.ToEnv())
	env = pkg.Merge(env, ruby.Integrations.ToEnv(ctx, diags))

	return env
}

func (ruby *Ruby) FromEnv(ctx context.Context, env *maps.Map[string, string], diags *diag.Diagnostics) {
	if ruby == nil || env == nil {
		return
	}

	pkg.ReadEnv(env, ruby, envVars...)
	ruby.StaticWebroot = pkg.FromStrPtr(env.PopPtr("STATIC_WEBROOT"))

	ruby.Integrations = attributes.FromEnvIntegrations(ctx, env, ruby.Integrations, diags)
}
