package php

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.dev/sdk/app"
)

var (
	PHPVersionEnv = pkg.NewStringEnv[PHP](app.CC_PHP_VERSION,
		func(p *PHP) types.String { return p.PHPVersion },
		func(p *PHP, v types.String) { p.PHPVersion = v })
	WebRootEnv = pkg.NewStringEnv[PHP](app.CC_WEBROOT,
		func(p *PHP) types.String { return p.WebRoot },
		func(p *PHP, v types.String) { p.WebRoot = v })
	RedisSessionsEnv = pkg.NewBoolEnv[PHP](app.SESSION_TYPE,
		func(p *PHP) string {
			if p.RedisSessions.ValueBool() {
				return "redis"
			}
			return ""
		},
		func(p *PHP, value string) {
			// an absent variable leaves the attribute as it is
			if value == "redis" {
				p.RedisSessions = types.BoolValue(true)
			}
		})
	DevDependenciesEnv = pkg.NewBoolEnv[PHP](app.CC_PHP_DEV_DEPENDENCIES,
		func(p *PHP) string {
			if p.DevDependencies.ValueBool() {
				return "install"
			}
			return ""
		},
		func(p *PHP, value string) {
			if value == "install" {
				p.DevDependencies = types.BoolValue(true)
			}
		})
	ComposerVersionEnv = pkg.NewStringEnv[PHP](app.CC_COMPOSER_VERSION,
		func(p *PHP) types.String { return p.ComposerVersion },
		func(p *PHP, v types.String) { p.ComposerVersion = v })
	CGIImplementationEnv = pkg.NewStringEnv[PHP](app.CC_CGI_IMPLEMENTATION,
		func(p *PHP) types.String { return p.CGIImplementation },
		func(p *PHP, v types.String) { p.CGIImplementation = v })
	ApacheHeadersSizeEnv = pkg.NewInt64Env[PHP](app.CC_APACHE_HEADERS_SIZE,
		func(p *PHP) types.Int64 { return p.ApacheHeadersSize },
		func(p *PHP, v types.Int64) { p.ApacheHeadersSize = v })
	HTTPTimeoutEnv = pkg.NewInt64Env[PHP](app.HTTP_TIMEOUT,
		func(p *PHP) types.Int64 { return p.HTTPTimeout },
		func(p *PHP, v types.Int64) { p.HTTPTimeout = v })
	MaxInputVarsEnv = pkg.NewInt64Env[PHP](app.MAX_INPUT_VARS,
		func(p *PHP) types.Int64 { return p.MaxInputVars },
		func(p *PHP, v types.Int64) { p.MaxInputVars = v })
	MemoryLimitEnv = pkg.NewInt64Env[PHP](app.MEMORY_LIMIT,
		func(p *PHP) types.Int64 { return p.MemoryLimit },
		func(p *PHP, v types.Int64) { p.MemoryLimit = v })
	RealpathCacheTTLEnv = pkg.NewInt64Env[PHP](app.CC_REALPATH_CACHE_TTL,
		func(p *PHP) types.Int64 { return p.RealpathCacheTTL },
		func(p *PHP, v types.Int64) { p.RealpathCacheTTL = v })
	OpcacheMemoryEnv = pkg.NewInt64Env[PHP](app.CC_OPCACHE_MEMORY,
		func(p *PHP) types.Int64 { return p.OpcacheMemory },
		func(p *PHP, v types.Int64) { p.OpcacheMemory = v })
	OpcacheMaxAcceleratedFilesEnv = pkg.NewInt64Env[PHP](app.CC_OPCACHE_MAX_ACCELERATED_FILES,
		func(p *PHP) types.Int64 { return p.OpcacheMaxAcceleratedFiles },
		func(p *PHP, v types.Int64) { p.OpcacheMaxAcceleratedFiles = v })
	OpcacheInternedStringsBufferEnv = pkg.NewInt64Env[PHP](app.CC_OPCACHE_INTERNED_STRINGS_BUFFER,
		func(p *PHP) types.Int64 { return p.OpcacheInternedStringsBuffer },
		func(p *PHP, v types.Int64) { p.OpcacheInternedStringsBuffer = v })
	OpcachePreloadEnv = pkg.NewStringEnv[PHP](app.CC_OPCACHE_PRELOAD,
		func(p *PHP) types.String { return p.OpcachePreload },
		func(p *PHP, v types.String) { p.OpcachePreload = v })
	AsyncAppBucketEnv = pkg.NewBoolEnv[PHP](app.CC_PHP_ASYNC_APP_BUCKET,
		func(p *PHP) string {
			if p.AsyncAppBucket.ValueBool() {
				return "async"
			}
			return ""
		},
		func(p *PHP, value string) {
			if value == "async" {
				p.AsyncAppBucket = types.BoolValue(true)
			}
		})
	DisableAppBucketEnv = pkg.NewBoolEnv[PHP](app.CC_PHP_DISABLE_APP_BUCKET,
		func(p *PHP) string {
			if p.DisableAppBucket.ValueBool() {
				return "true"
			}
			return ""
		},
		func(p *PHP, value string) {
			if value == "true" || value == "yes" || value == "disable" {
				p.DisableAppBucket = types.BoolValue(true)
			}
		})
	APMAgentEnabledEnv = pkg.NewBoolEnv[PHP](app.ENABLE_ELASTIC_APM_AGENT,
		func(p *PHP) string {
			// false is written too: the platform enables the agent on its own
			// when an Elastic APM server is configured
			if p.APMAgentEnabled.IsNull() || p.APMAgentEnabled.IsUnknown() {
				return ""
			}
			return strconv.FormatBool(p.APMAgentEnabled.ValueBool())
		},
		func(p *PHP, value string) {
			switch value {
			case "true":
				p.APMAgentEnabled = types.BoolValue(true)
			case "false":
				p.APMAgentEnabled = types.BoolValue(false)
			}
		})
)

// envVars lists every php attribute backed by an environment variable.
var envVars = []pkg.EnvVarMapping[*PHP]{
	application.OnRuntime[*PHP](application.AppFolderEnv),
	PHPVersionEnv,
	WebRootEnv,
	RedisSessionsEnv,
	DevDependenciesEnv,
	ComposerVersionEnv,
	CGIImplementationEnv,
	ApacheHeadersSizeEnv,
	HTTPTimeoutEnv,
	MaxInputVarsEnv,
	MemoryLimitEnv,
	RealpathCacheTTLEnv,
	OpcacheMemoryEnv,
	OpcacheMaxAcceleratedFilesEnv,
	OpcacheInternedStringsBufferEnv,
	OpcachePreloadEnv,
	AsyncAppBucketEnv,
	DisableAppBucketEnv,
	APMAgentEnabledEnv,
}
