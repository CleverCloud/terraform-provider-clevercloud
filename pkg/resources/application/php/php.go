package php

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
)

type ResourcePHP struct {
	helper.Configurer
}

const (
	ALWAYS_POPULATE_RAW_POST_DATA      = "ALWAYS_POPULATE_RAW_POST_DATA"
	CC_COMPOSER_VERSION                = "CC_COMPOSER_VERSION"
	CC_CGI_IMPLEMENTATION              = "CC_CGI_IMPLEMENTATION"
	CC_HTTP_BASIC_AUTH                 = "CC_HTTP_BASIC_AUTH"
	CC_APACHE_HEADERS_SIZE             = "CC_APACHE_HEADERS_SIZE"
	CC_LDAP_CA_CERT                    = "CC_LDAP_CA_CERT"
	CC_MTA_AUTH_PASSWORD               = "CC_MTA_AUTH_PASSWORD"
	CC_MTA_AUTH_USER                   = "CC_MTA_AUTH_USER"
	CC_MTA_SERVER_AUTH_METHOD          = "CC_MTA_SERVER_AUTH_METHOD"
	CC_MTA_SERVER_HOST                 = "CC_MTA_SERVER_HOST"
	CC_MTA_SERVER_PORT                 = "CC_MTA_SERVER_PORT"
	CC_MTA_SERVER_USE_TLS              = "CC_MTA_SERVER_USE_TLS"
	CC_OPCACHE_INTERNED_STRINGS_BUFFER = "CC_OPCACHE_INTERNED_STRINGS_BUFFER"
	CC_OPCACHE_MAX_ACCELERATED_FILES   = "CC_OPCACHE_MAX_ACCELERATED_FILES"
	CC_OPCACHE_MEMORY                  = "CC_OPCACHE_MEMORY"
	CC_OPCACHE_PRELOAD                 = "CC_OPCACHE_PRELOAD"
	CC_PHP_ASYNC_APP_BUCKET            = "CC_PHP_ASYNC_APP_BUCKET"
	CC_PHP_DEV_DEPENDENCIES            = "CC_PHP_DEV_DEPENDENCIES"
	CC_PHP_DISABLE_APP_BUCKET          = "CC_PHP_DISABLE_APP_BUCKET"
	CC_PHP_VERSION                     = "CC_PHP_VERSION"
	CC_REALPATH_CACHE_TTL              = "CC_REALPATH_CACHE_TTL"
	CC_WEBROOT                         = "CC_WEBROOT"
	ENABLE_ELASTIC_APM_AGENT           = "ENABLE_ELASTIC_APM_AGENT"
	ENABLE_GRPC                        = "ENABLE_GRPC"
	ENABLE_PDFLIB                      = "ENABLE_PDFLIB"
	ENABLE_REDIS                       = "ENABLE_REDIS"
	HTTP_TIMEOUT                       = "HTTP_TIMEOUT"
	LDAPTLS_CACERT                     = "LDAPTLS_CACERT"
	MAX_INPUT_VARS                     = "MAX_INPUT_VARS"
	MEMORY_LIMIT                       = "MEMORY_LIMIT"
	SESSION_TYPE                       = "SESSION_TYPE"
	SOCKSIFY_EVERYTHING                = "SOCKSIFY_EVERYTHING"
	SQREEN_API_APP_NAME                = "SQREEN_API_APP_NAME"
	SQREEN_API_TOKEN                   = "SQREEN_API_TOKEN"
)

func NewResourcePHP() resource.Resource {
	return &ResourcePHP{}
}

func (r *ResourcePHP) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_php"
}

// UpgradeState implements state migration from version 0 to 1 for vhosts attribute
func (r *ResourcePHP) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaPHPV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				tflog.Info(ctx, "Upgrading PHP resource state from version 0 to 1")

				old := helper.StateFrom[PHPV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				oldVhosts := []string{}
				res.Diagnostics.Append(old.VHosts.ElementsAs(ctx, &oldVhosts, false)...)
				vhosts := helper.VHostsFromAPIHosts(ctx, oldVhosts, old.VHosts, &res.Diagnostics)

				newState := PHP{
					Runtime: application.Runtime{
						ID:               old.ID,
						Name:             old.Name,
						Description:      old.Description,
						MinInstanceCount: old.MinInstanceCount,
						MaxInstanceCount: old.MaxInstanceCount,
						SmallestFlavor:   old.SmallestFlavor,
						BiggestFlavor:    old.BiggestFlavor,
						BuildFlavor:      old.BuildFlavor,
						Region:           old.Region,
						StickySessions:   old.StickySessions,
						RedirectHTTPS:    old.RedirectHTTPS,
						VHosts:           vhosts,
						DeployURL:        old.DeployURL,
						Dependencies:     old.Dependencies,
						Deployment:       old.Deployment,
						Hooks:            old.Hooks,
						AppFolder:        old.AppFolder,
						Environment:      old.Environment,
					},
					PHPVersion: old.PHPVersion,
					WebRoot:    old.WebRoot,
				}

				// Migrate RedisSessions (Bool) to SessionType (String)
				if !old.RedisSessions.IsNull() && old.RedisSessions.ValueBool() {
					newState.SessionType = pkg.FromStr("redis")
				}

				// Migrate DevDependencies from Bool to String
				if !old.DevDependencies.IsNull() {
					if old.DevDependencies.ValueBool() {
						newState.DevDependencies = pkg.FromStr("install")
					} else {
						newState.DevDependencies = pkg.FromStr("ignore")
					}
				}

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}
