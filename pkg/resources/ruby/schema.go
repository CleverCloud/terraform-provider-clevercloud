package ruby

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

type Ruby struct {
	attributes.Runtime
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
	attributes.RuntimeV0
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
	Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
		// CC_RUBY_VERSION
		"ruby_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Ruby version to use (e.g., '3.3', '3.3.1')",
		},
		// CC_ENABLE_SIDEKIQ
		"enable_sidekiq": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
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
			Sensitive:           true,
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
			Computed:            true,
			Default:             booldefault.StaticBool(false),
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

var schemaRubyV0 = schema.Schema{
	Version:             0,
	MarkdownDescription: rubyDoc,
	Attributes: attributes.WithRuntimeCommonsV0(map[string]schema.Attribute{
		// CC_RUBY_VERSION
		"ruby_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Ruby version to use (e.g., '3.3', '3.3.1')",
		},
		// CC_ENABLE_SIDEKIQ
		"enable_sidekiq": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
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
			Sensitive:           true,
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
			Computed:            true,
			Default:             booldefault.StaticBool(false),
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

func (ruby Ruby) toEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	env := ruby.ToEnv(ctx, diags)
	if diags.HasError() {
		return env
	}

	pkg.IfIsSetStr(ruby.RubyVersion, func(s string) { env[CC_RUBY_VERSION] = s })
	pkg.IfIsSetB(ruby.EnableSidekiq, func(b bool) { env[CC_ENABLE_SIDEKIQ] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(ruby.RackupServer, func(s string) { env[CC_RACKUP_SERVER] = s })
	pkg.IfIsSetStr(ruby.RakeGoals, func(s string) { env[CC_RAKEGOALS] = s })
	pkg.IfIsSetStr(ruby.SidekiqFiles, func(s string) { env[CC_SIDEKIQ_FILES] = s })
	pkg.IfIsSetStr(ruby.HTTPBasicAuth, func(s string) { env[CC_HTTP_BASIC_AUTH] = s })
	pkg.IfIsSetStr(ruby.NginxProxyBuffers, func(s string) { env[CC_NGINX_PROXY_BUFFERS] = s })
	pkg.IfIsSetStr(ruby.NginxProxyBufferSize, func(s string) { env[CC_NGINX_PROXY_BUFFER_SIZE] = s })
	pkg.IfIsSetB(ruby.EnableGzipCompression, func(b bool) { env[ENABLE_GZIP_COMPRESSION] = strconv.FormatBool(b) })
	pkg.IfIsSetStr(ruby.GzipTypes, func(s string) { env[GZIP_TYPES] = s })
	pkg.IfIsSetI(ruby.NginxReadTimeout, func(i int64) { env[NGINX_READ_TIMEOUT] = strconv.FormatInt(i, 10) })
	pkg.IfIsSetStr(ruby.RackEnv, func(s string) { env[RACK_ENV] = s })
	pkg.IfIsSetStr(ruby.RailsEnv, func(s string) { env[RAILS_ENV] = s })
	pkg.IfIsSetStr(ruby.StaticFilesPath, func(s string) { env[STATIC_FILES_PATH] = s })
	pkg.IfIsSetStr(ruby.StaticURLPrefix, func(s string) { env[STATIC_URL_PREFIX] = s })
	pkg.IfIsSetStr(ruby.StaticWebroot, func(s string) { env[STATIC_WEBROOT] = s })

	return env
}

func (ruby *Ruby) fromEnv(ctx context.Context, env map[string]string) {
	m := helper.NewEnvMap(env)

	ruby.RubyVersion = pkg.FromStr(m.Pop(CC_RUBY_VERSION))
	if sidekiq, err := strconv.ParseBool(m.Pop(CC_ENABLE_SIDEKIQ)); err == nil {
		ruby.EnableSidekiq = pkg.FromBool(sidekiq)
	}
	ruby.RackupServer = pkg.FromStr(m.Pop(CC_RACKUP_SERVER))
	ruby.RakeGoals = pkg.FromStr(m.Pop(CC_RAKEGOALS))
	ruby.SidekiqFiles = pkg.FromStr(m.Pop(CC_SIDEKIQ_FILES))
	ruby.HTTPBasicAuth = pkg.FromStr(m.Pop(CC_HTTP_BASIC_AUTH))
	ruby.NginxProxyBuffers = pkg.FromStr(m.Pop(CC_NGINX_PROXY_BUFFERS))
	ruby.NginxProxyBufferSize = pkg.FromStr(m.Pop(CC_NGINX_PROXY_BUFFER_SIZE))
	if gzip, err := strconv.ParseBool(m.Pop(ENABLE_GZIP_COMPRESSION)); err == nil {
		ruby.EnableGzipCompression = pkg.FromBool(gzip)
	}
	ruby.GzipTypes = pkg.FromStr(m.Pop(GZIP_TYPES))
	if timeout, err := strconv.ParseInt(m.Pop(NGINX_READ_TIMEOUT), 10, 64); err == nil {
		ruby.NginxReadTimeout = pkg.FromI(timeout)
	}
	ruby.RackEnv = pkg.FromStr(m.Pop(RACK_ENV))
	ruby.RailsEnv = pkg.FromStr(m.Pop(RAILS_ENV))
	ruby.StaticFilesPath = pkg.FromStr(m.Pop(STATIC_FILES_PATH))
	ruby.StaticURLPrefix = pkg.FromStr(m.Pop(STATIC_URL_PREFIX))
	ruby.StaticWebroot = pkg.FromStr(m.Pop(STATIC_WEBROOT))

	ruby.FromEnvironment(ctx, m)
}

func (ruby Ruby) toDeployment(gitAuth *http.BasicAuth) *application.Deployment {
	if ruby.Deployment == nil || ruby.Deployment.Repository.IsNull() {
		return nil
	}

	return &application.Deployment{
		Repository:    ruby.Deployment.Repository.ValueString(),
		Commit:        ruby.Deployment.Commit.ValueStringPointer(),
		CleverGitAuth: gitAuth,
	}
}
