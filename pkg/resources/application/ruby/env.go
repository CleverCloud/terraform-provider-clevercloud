package ruby

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

// Variables read by the Ruby application itself, which the platform does not
// list among its own.
const (
	rackEnv  = "RACK_ENV"
	railsEnv = "RAILS_ENV"
)

// trueToEnv writes "true" when the attribute is true, nothing otherwise.
func trueToEnv(b types.Bool) string {
	if b.ValueBool() {
		return "true"
	}
	return ""
}

var (
	RubyVersionEnv = pkg.NewStringEnv[Ruby](app.CC_RUBY_VERSION,
		func(r *Ruby) types.String { return r.RubyVersion },
		func(r *Ruby, v types.String) { r.RubyVersion = v })
	EnableSidekiqEnv = pkg.NewBoolEnv[Ruby](app.CC_ENABLE_SIDEKIQ,
		func(r *Ruby) string { return trueToEnv(r.EnableSidekiq) },
		func(r *Ruby, value string) {
			// an absent variable leaves the attribute as it is
			if value == "true" {
				r.EnableSidekiq = types.BoolValue(true)
			}
		})
	RackupServerEnv = pkg.NewStringEnv[Ruby](app.CC_RACKUP_SERVER,
		func(r *Ruby) types.String { return r.RackupServer },
		func(r *Ruby, v types.String) { r.RackupServer = v })
	RakeGoalsEnv = pkg.NewStringEnv[Ruby](app.CC_RAKEGOALS,
		func(r *Ruby) types.String { return r.RakeGoals },
		func(r *Ruby, v types.String) { r.RakeGoals = v })
	SidekiqFilesEnv = pkg.NewStringEnv[Ruby](app.CC_SIDEKIQ_FILES,
		func(r *Ruby) types.String { return r.SidekiqFiles },
		func(r *Ruby, v types.String) { r.SidekiqFiles = v })
	HTTPBasicAuthEnv = pkg.NewStringEnv[Ruby](app.CC_HTTP_BASIC_AUTH,
		func(r *Ruby) types.String { return r.HTTPBasicAuth },
		func(r *Ruby, v types.String) { r.HTTPBasicAuth = v })
	NginxProxyBuffersEnv = pkg.NewStringEnv[Ruby](app.CC_NGINX_PROXY_BUFFERS,
		func(r *Ruby) types.String { return r.NginxProxyBuffers },
		func(r *Ruby, v types.String) { r.NginxProxyBuffers = v })
	NginxProxyBufferSizeEnv = pkg.NewStringEnv[Ruby](app.CC_NGINX_PROXY_BUFFER_SIZE,
		func(r *Ruby) types.String { return r.NginxProxyBufferSize },
		func(r *Ruby, v types.String) { r.NginxProxyBufferSize = v })
	EnableGzipCompressionEnv = pkg.NewBoolEnv[Ruby](app.ENABLE_GZIP_COMPRESSION,
		func(r *Ruby) string { return trueToEnv(r.EnableGzipCompression) },
		func(r *Ruby, value string) {
			if value == "true" {
				r.EnableGzipCompression = types.BoolValue(true)
			}
		})
	GzipTypesEnv = pkg.NewStringEnv[Ruby](app.GZIP_TYPES,
		func(r *Ruby) types.String { return r.GzipTypes },
		func(r *Ruby, v types.String) { r.GzipTypes = v })
	NginxReadTimeoutEnv = pkg.NewInt64Env[Ruby](app.NGINX_READ_TIMEOUT,
		func(r *Ruby) types.Int64 { return r.NginxReadTimeout },
		func(r *Ruby, v types.Int64) { r.NginxReadTimeout = v })
	RackEnvEnv = pkg.NewStringEnv[Ruby](rackEnv,
		func(r *Ruby) types.String { return r.RackEnv },
		func(r *Ruby, v types.String) { r.RackEnv = v })
	RailsEnvEnv = pkg.NewStringEnv[Ruby](railsEnv,
		func(r *Ruby) types.String { return r.RailsEnv },
		func(r *Ruby, v types.String) { r.RailsEnv = v })
	StaticFilesPathEnv = pkg.NewStringEnv[Ruby](app.STATIC_FILES_PATH,
		func(r *Ruby) types.String { return r.StaticFilesPath },
		func(r *Ruby, v types.String) { r.StaticFilesPath = v })
	StaticURLPrefixEnv = pkg.NewStringEnv[Ruby](app.STATIC_URL_PREFIX,
		func(r *Ruby) types.String { return r.StaticURLPrefix },
		func(r *Ruby, v types.String) { r.StaticURLPrefix = v })
)

// envVars lists every ruby attribute backed by an environment variable.
// The deprecated static_webroot is left out: the platform ignores its
// variable, so environment may still hold it.
var envVars = []pkg.EnvVarMapping[*Ruby]{
	application.OnRuntime[*Ruby](application.AppFolderEnv),
	RubyVersionEnv,
	EnableSidekiqEnv,
	RackupServerEnv,
	RakeGoalsEnv,
	SidekiqFilesEnv,
	HTTPBasicAuthEnv,
	NginxProxyBuffersEnv,
	NginxProxyBufferSizeEnv,
	EnableGzipCompressionEnv,
	GzipTypesEnv,
	NginxReadTimeoutEnv,
	RackEnvEnv,
	RailsEnvEnv,
	StaticFilesPathEnv,
	StaticURLPrefixEnv,
}
