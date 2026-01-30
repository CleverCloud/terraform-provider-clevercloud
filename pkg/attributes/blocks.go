package attributes

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	helperMaps "github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
)

// used to identify when the repo deployment is handled by a Github hook
// In this case, we must not deploy with Terraform
const GITHUB_COMMIT_PREFIX = "github_hook"

const (
	CC_REDIRECTIONIO_PROJECT_KEY   = "CC_REDIRECTIONIO_PROJECT_KEY"
	CC_REDIRECTIONIO_INSTANCE_NAME = "CC_REDIRECTIONIO_INSTANCE_NAME"
	CC_REDIRECTIONIO_BACKEND_PORT  = "CC_REDIRECTIONIO_BACKEND_PORT"

	NEW_RELIC_LICENSE_KEY = "NEW_RELIC_LICENSE_KEY"
	NEW_RELIC_APP_NAME    = "NEW_RELIC_APP_NAME"

	CC_CLAMAV = "CC_CLAMAV"

	CC_VARNISH_FILE         = "CC_VARNISH_FILE"
	CC_VARNISH_STORAGE_SIZE = "CC_VARNISH_STORAGE_SIZE"

	CC_METRICS_PROMETHEUS_USER             = "CC_METRICS_PROMETHEUS_USER"
	CC_METRICS_PROMETHEUS_PASSWORD         = "CC_METRICS_PROMETHEUS_PASSWORD"
	CC_METRICS_PROMETHEUS_PATH             = "CC_METRICS_PROMETHEUS_PATH"
	CC_METRICS_PROMETHEUS_PORT             = "CC_METRICS_PROMETHEUS_PORT"
	CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT = "CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT"

	CC_ENABLE_PGPOOL = "CC_ENABLE_PGPOOL"
)

// Deployment block
type Deployment struct {
	Repository          types.String `tfsdk:"repository"`
	Commit              types.String `tfsdk:"commit"`
	BasicAuthentication types.String `tfsdk:"authentication_basic"`
}

// Hooks block
type Hooks struct {
	PreBuild   types.String `tfsdk:"pre_build"`
	PostBuild  types.String `tfsdk:"post_build"`
	PreRun     types.String `tfsdk:"pre_run"`
	RunSucceed types.String `tfsdk:"run_succeed"`
	RunFailed  types.String `tfsdk:"run_failed"`
}

type Redirectionio struct {
	ProjectKey   types.String `tfsdk:"project_key"`
	InstanceName types.String `tfsdk:"instance_name"`
	BackendPort  types.Int64  `tfsdk:"backend_port"`
}

var RedirectionIOSchema = map[string]attr.Type{
	"project_key":   types.StringType,
	"instance_name": types.StringType,
	"backend_port":  types.Int64Type,
}

// NewRelic integration configuration
type NewRelic struct {
	LicenseKey types.String `tfsdk:"license_key"`
	AppName    types.String `tfsdk:"app_name"`
}

var NewRelicSchema = map[string]attr.Type{
	"license_key": types.StringType,
	"app_name":    types.StringType,
}

// ClamAV integration configuration
type ClamAV struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

var ClamavSchema = map[string]attr.Type{
	"enabled": types.BoolType,
}

// Varnish integration configuration
type Varnish struct {
	ConfigFile  types.String `tfsdk:"config_file"`
	StorageSize types.String `tfsdk:"storage_size"`
}

var VarnishSchema = map[string]attr.Type{
	"config_file":  types.StringType,
	"storage_size": types.StringType,
}

// Prometheus integration configuration
type Prometheus struct {
	User            types.String `tfsdk:"user"`
	Password        types.String `tfsdk:"password"`
	Path            types.String `tfsdk:"path"`
	Port            types.Int64  `tfsdk:"port"`
	ResponseTimeout types.Int64  `tfsdk:"response_timeout"`
}

var PrometheusSchema = map[string]attr.Type{
	"user":             types.StringType,
	"password":         types.StringType,
	"path":             types.StringType,
	"port":             types.Int64Type,
	"response_timeout": types.Int64Type,
}

// PgPool-II integration configuration
type PgPoolII struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

var PGPoolIISchema = map[string]attr.Type{
	"enabled": types.BoolType,
}

type Integrations struct {
	Redirectionio types.Object `tfsdk:"redirectionio"`
	NewRelic      types.Object `tfsdk:"newrelic"`
	ClamAV        types.Object `tfsdk:"clamav"`
	Varnish       types.Object `tfsdk:"varnish"`
	Prometheus    types.Object `tfsdk:"prometheus"`
	PgPoolII      types.Object `tfsdk:"pgpoolii"`
}

var IntegrationsAttribute = schema.SingleNestedAttribute{
	Optional:            true,
	MarkdownDescription: "Third-party integrations configuration",
	Attributes: map[string]schema.Attribute{
		"redirectionio": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Redirection.io integration for URL redirection and traffic management. See [Redirection.io docs](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#redirectionio)",
			Attributes: map[string]schema.Attribute{
				"project_key": schema.StringAttribute{
					Required:            true,
					Sensitive:           true,
					MarkdownDescription: "Your Redirection.io project key ([CC_REDIRECTIONIO_PROJECT_KEY](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#redirectionio))",
				},
				"instance_name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Custom instance name for the Redirection.io agent ([CC_REDIRECTIONIO_INSTANCE_NAME](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#redirectionio))",
				},
				"backend_port": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Backend application port ([CC_REDIRECTIONIO_BACKEND_PORT](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#redirectionio))",
				},
			},
		},
		"newrelic": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "New Relic APM integration for application performance monitoring",
			Attributes: map[string]schema.Attribute{
				"license_key": schema.StringAttribute{
					Required:            true,
					Sensitive:           true,
					MarkdownDescription: "Your New Relic license key ([NEW_RELIC_LICENSE_KEY](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#new-relic))",
				},
				"app_name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Application name in New Relic ([NEW_RELIC_APP_NAME](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#new-relic))",
				},
			},
		},
		"clamav": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "ClamAV antivirus integration for file scanning",
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "Enable ClamAV antivirus ([CC_CLAMAV](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#clamav))",
				},
			},
		},
		"varnish": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Varnish cache integration for HTTP acceleration. See [Varnish docs](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#varnish)",
			Attributes: map[string]schema.Attribute{
				"config_file": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Path to the Varnish configuration file, relative to your application root ([CC_VARNISH_FILE](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#varnish)). Default: `/clevercloud/varnish.vcl`",
				},
				"storage_size": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Configure the size of the Varnish cache ([CC_VARNISH_STORAGE_SIZE](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#varnish)). Default: `1G`",
				},
			},
		},
		"prometheus": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Prometheus metrics integration. See [Prometheus docs](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#prometheus)",
			Attributes: map[string]schema.Attribute{
				"user": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Define the user for the basic auth of the Prometheus endpoint ([CC_METRICS_PROMETHEUS_USER](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#prometheus))",
				},
				"password": schema.StringAttribute{
					Optional:            true,
					Sensitive:           true,
					MarkdownDescription: "Define the password for the basic auth of the Prometheus endpoint ([CC_METRICS_PROMETHEUS_PASSWORD](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#prometheus))",
				},
				"path": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Define the path on which the Prometheus endpoint is available ([CC_METRICS_PROMETHEUS_PATH](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#prometheus)). Default: `/metrics`",
				},
				"port": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Define the port on which the Prometheus endpoint is available ([CC_METRICS_PROMETHEUS_PORT](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#prometheus)). Default: `9100`",
				},
				"response_timeout": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Define the timeout in seconds to collect the application metrics. This value must be below 60 seconds ([CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#prometheus)). Default: `3`",
				},
			},
		},
		"pgpoolii": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "PgPool-II configuration. [Learn more](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#pgpool-ii)",
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "Enable PgPool-II connection pooler ([CC_ENABLE_PGPOOL](https://www.clever.cloud/developers/doc/reference/reference-environment-variables#pgpool-ii))",
				},
			},
		},
	},
}

var blocks = map[string]schema.Block{
	"deployment": schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				Optional:            true, // If "deployment" attribute is defined, then repository is required
				Description:         "The repository URL to deploy, can be 'https://...', 'file://...'",
				MarkdownDescription: "The repository URL to deploy, can be 'https://...', 'file://...'",
			},
			"commit": schema.StringAttribute{
				Optional:            true,
				Description:         "The git reference you want to deploy",
				MarkdownDescription: "Support multiple syntax like `refs/heads/[BRANCH]`, `github_hook` or `[COMMIT]`, when using the special value `github_hook`, we will link the application to the Github repository",
				Validators: []validator.String{
					pkg.NewValidator(
						"if reference (not commit hash) is provided test it's syntax",
						func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() || !strings.Contains(req.ConfigValue.ValueString(), "/") {
								return
							}

							if req.ConfigValue.ValueString() == GITHUB_COMMIT_PREFIX {
								return
							}

							ref := plumbing.ReferenceName(req.ConfigValue.ValueString())
							if err := ref.Validate(); err != nil {
								res.Diagnostics.AddAttributeError(
									req.Path,
									"invalid Git reference",
									err.Error(),
								)
							}
						}),
				},
			},
			"authentication_basic": schema.StringAttribute{
				Sensitive: true,
				Optional:  true,
				// TODO: investigate behaviour (not available in plan, but available in config ?)
				//WriteOnly:           true,
				MarkdownDescription: "user ans password ':' separated, (PersonalAccessToken in Github case)",
				Validators:          []validator.String{UserPasswordInput},
			},
		},
	},
	"hooks": schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"pre_build": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook is ran before the dependencies are fetched. If it fails, the deployment fails",
				MarkdownDescription: "[CC_PRE_BUILD_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#pre-build)",
			},
			"post_build": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook is ran after the project is built, and before the cache archive is generated. If it fails, the deployment fails",
				MarkdownDescription: "[CC_POST_BUILD_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#post-build)",
			},
			"pre_run": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook is ran before the application is started, but after the cache archive has been generated. If it fails, the deployment fails.",
				MarkdownDescription: "[CC_PRE_RUN_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#pre-run)",
			},
			"run_succeed": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook are ran once the application has started. Their failure doesn't cause the deployment to fail.",
				MarkdownDescription: "[CC_RUN_SUCCEEDED_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#run-successfail)",
			},
			"run_failed": schema.StringAttribute{
				Optional:            true,
				Description:         "This hook are ran once the application has failed starting.",
				MarkdownDescription: "[CC_RUN_FAILED_HOOK](https://www.clever.cloud/developers/doc/develop/build-hooks/#run-successfail)",
			},
		},
	},
}

var UserPasswordInput = pkg.NewStringValidator("format must be 'user:password'", func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v := req.ConfigValue.ValueString()
	splits := strings.SplitN(v, ":", 2)
	if len(splits) != 2 {
		res.Diagnostics.AddError("invalid authenication value", "expect user:password")
		return
	}
	if splits[0] == "" && splits[1] == "" {
		res.Diagnostics.AddError("invalid authenication value", "user and password cannot be both empty")
	}
})

func WithBlockRuntimeCommons(runtimeSpecifics map[string]schema.Block) map[string]schema.Block {
	m := map[string]schema.Block{}
	maps.Copy(m, blocks)
	maps.Copy(m, runtimeSpecifics)
	return m
}

func (hooks *Hooks) ToEnv() map[string]string {
	m := map[string]string{}

	if hooks == nil {
		return m
	}

	pkg.IfIsSetStr(hooks.PreBuild, func(script string) { m["CC_PRE_BUILD_HOOK"] = script })
	pkg.IfIsSetStr(hooks.PostBuild, func(script string) { m["CC_POST_BUILD_HOOK"] = script })
	pkg.IfIsSetStr(hooks.PreRun, func(script string) { m["CC_PRE_RUN_HOOK"] = script })
	pkg.IfIsSetStr(hooks.RunFailed, func(script string) { m["CC_RUN_FAILED_HOOK"] = script })
	pkg.IfIsSetStr(hooks.RunSucceed, func(script string) { m["CC_RUN_SUCCEEDED_HOOK"] = script })

	return m
}

// FromEnvHooks extracts hook variables from environment and removes them from the map
func FromEnvHooks(env *helperMaps.Map[string, string], oldValue *Hooks) *Hooks {
	hasPreBuild := env.Has("CC_PRE_BUILD_HOOK")
	hasPostBuild := env.Has("CC_POST_BUILD_HOOK")
	hasPreRun := env.Has("CC_PRE_RUN_HOOK")
	hasRunFailed := env.Has("CC_RUN_FAILED_HOOK")
	hasRunSucceed := env.Has("CC_RUN_SUCCEEDED_HOOK")

	hasAnyHook := hasPreBuild || hasPostBuild || hasPreRun || hasRunFailed || hasRunSucceed

	if !hasAnyHook {
		// No hooks in API
		if oldValue == nil {
			// User never specified hooks
			return nil
		}
		// User had hooks, but they're gone from API - return empty hooks block
	}

	hooks := &Hooks{
		PreBuild:   pkg.FromStrPtr(env.PopPtr("CC_PRE_BUILD_HOOK")),
		PostBuild:  pkg.FromStrPtr(env.PopPtr("CC_POST_BUILD_HOOK")),
		PreRun:     pkg.FromStrPtr(env.PopPtr("CC_PRE_RUN_HOOK")),
		RunFailed:  pkg.FromStrPtr(env.PopPtr("CC_RUN_FAILED_HOOK")),
		RunSucceed: pkg.FromStrPtr(env.PopPtr("CC_RUN_SUCCEEDED_HOOK")),
	}

	return hooks
}

func (integrations *Integrations) ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string {
	m := map[string]string{}
	if integrations == nil {
		return m
	}

	pkg.IfSetObject(ctx, diags, integrations.Redirectionio, func(redirio Redirectionio) {
		pkg.IfIsSetStr(redirio.ProjectKey, func(s string) {
			m[CC_REDIRECTIONIO_PROJECT_KEY] = s
		})
		pkg.IfIsSetStr(redirio.InstanceName, func(s string) {
			m[CC_REDIRECTIONIO_INSTANCE_NAME] = s
		})
		pkg.IfIsSetI(redirio.BackendPort, func(i int64) {
			m[CC_REDIRECTIONIO_BACKEND_PORT] = fmt.Sprintf("%d", i)
		})
	})

	pkg.IfSetObject(ctx, diags, integrations.NewRelic, func(newrelic NewRelic) {
		pkg.IfIsSetStr(newrelic.LicenseKey, func(s string) {
			m[NEW_RELIC_LICENSE_KEY] = s
		})
		pkg.IfIsSetStr(newrelic.AppName, func(s string) {
			m[NEW_RELIC_APP_NAME] = s
		})
	})

	pkg.IfSetObject(ctx, diags, integrations.ClamAV, func(clamav ClamAV) {
		pkg.IfIsSetB(clamav.Enabled, func(enabled bool) {
			if enabled {
				m[CC_CLAMAV] = "true"
			}
		})
	})

	pkg.IfSetObject(ctx, diags, integrations.Varnish, func(varnish Varnish) {
		pkg.IfIsSetStr(varnish.ConfigFile, func(s string) {
			m[CC_VARNISH_FILE] = s
		})
		pkg.IfIsSetStr(varnish.StorageSize, func(s string) {
			m[CC_VARNISH_STORAGE_SIZE] = s
		})
	})

	pkg.IfSetObject(ctx, diags, integrations.Prometheus, func(prom Prometheus) {
		pkg.IfIsSetStr(prom.User, func(s string) {
			m[CC_METRICS_PROMETHEUS_USER] = s
		})
		pkg.IfIsSetStr(prom.Password, func(s string) {
			m[CC_METRICS_PROMETHEUS_PASSWORD] = s
		})
		pkg.IfIsSetStr(prom.Path, func(s string) {
			m[CC_METRICS_PROMETHEUS_PATH] = s
		})
		pkg.IfIsSetI(prom.Port, func(i int64) {
			m[CC_METRICS_PROMETHEUS_PORT] = fmt.Sprintf("%d", i)
		})
		pkg.IfIsSetI(prom.ResponseTimeout, func(i int64) {
			m[CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT] = fmt.Sprintf("%d", i)
		})
	})

	pkg.IfSetObject(ctx, diags, integrations.PgPoolII, func(pgpool PgPoolII) {
		pkg.IfIsSetB(pgpool.Enabled, func(enabled bool) {
			if enabled {
				m[CC_ENABLE_PGPOOL] = "true"
			}
		})
	})

	return m
}

// FromEnvIntegrations syncs Integrations block with environment variables from API
// Logic:
// - If any integration env var exists: create integrations block and populate sub-blocks
// - If no integration env vars exist:
//   - If integrations block exists in state: clear all fields (drift detection)
//   - If integrations block doesn't exist: do nothing (stays nil)
func FromEnvIntegrations(ctx context.Context, env *helperMaps.Map[string, string], oldValue *Integrations, diags *diag.Diagnostics) *Integrations {
	hasRedirectionio := env.Has(CC_REDIRECTIONIO_PROJECT_KEY) ||
		env.Has(CC_REDIRECTIONIO_INSTANCE_NAME) ||
		env.Has(CC_REDIRECTIONIO_BACKEND_PORT)

	hasNewRelic := env.Has(NEW_RELIC_LICENSE_KEY) ||
		env.Has(NEW_RELIC_APP_NAME)

	hasClamAV := env.Has(CC_CLAMAV)

	hasVarnish := env.Has(CC_VARNISH_FILE) ||
		env.Has(CC_VARNISH_STORAGE_SIZE)

	hasPrometheus := env.Has(CC_METRICS_PROMETHEUS_USER) ||
		env.Has(CC_METRICS_PROMETHEUS_PASSWORD) ||
		env.Has(CC_METRICS_PROMETHEUS_PATH) ||
		env.Has(CC_METRICS_PROMETHEUS_PORT) ||
		env.Has(CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT)

	hasPgPoolII := env.Has(CC_ENABLE_PGPOOL)

	hasAnyIntegration := hasRedirectionio || hasNewRelic || hasClamAV || hasVarnish || hasPrometheus || hasPgPoolII

	integration := &Integrations{
		Redirectionio: types.ObjectNull(RedirectionIOSchema),
		NewRelic:      types.ObjectNull(NewRelicSchema),
		ClamAV:        types.ObjectNull(ClamavSchema),
		Varnish:       types.ObjectNull(VarnishSchema),
		Prometheus:    types.ObjectNull(PrometheusSchema),
		PgPoolII:      types.ObjectNull(PGPoolIISchema),
	}

	if !hasAnyIntegration { // nothing on API side
		if oldValue == nil { // whether practitioner has set "integration" attribute
			return nil
		}
	}

	var d diag.Diagnostics
	var obj basetypes.ObjectValue

	// Populate each integration based on its env vars
	if hasRedirectionio {
		redir := Redirectionio{
			ProjectKey:   pkg.FromStrPtr(env.PopPtr(CC_REDIRECTIONIO_PROJECT_KEY)),
			InstanceName: pkg.FromStrPtr(env.PopPtr(CC_REDIRECTIONIO_INSTANCE_NAME)),
			BackendPort:  pkg.FromIntPtr(env.PopPtr(CC_REDIRECTIONIO_BACKEND_PORT)),
		}
		obj, d = types.ObjectValueFrom(ctx, RedirectionIOSchema, redir)
		integration.Redirectionio = obj
	}

	if hasNewRelic {
		nr := NewRelic{
			LicenseKey: pkg.FromStrPtr(env.PopPtr(NEW_RELIC_LICENSE_KEY)),
			AppName:    pkg.FromStrPtr(env.PopPtr(NEW_RELIC_APP_NAME)),
		}
		obj, d = types.ObjectValueFrom(ctx, NewRelicSchema, nr)
		integration.NewRelic = obj
	}

	if hasClamAV {
		clamav := ClamAV{
			Enabled: pkg.FromBoolPtr(env.PopPtr(CC_CLAMAV)),
		}
		obj, d = types.ObjectValueFrom(ctx, ClamavSchema, clamav)
		integration.ClamAV = obj
	}

	if hasVarnish {
		varnish := Varnish{
			ConfigFile:  pkg.FromStrPtr(env.PopPtr(CC_VARNISH_FILE)),
			StorageSize: pkg.FromStrPtr(env.PopPtr(CC_VARNISH_STORAGE_SIZE)),
		}
		obj, d = types.ObjectValueFrom(ctx, VarnishSchema, varnish)
		integration.Varnish = obj
	}

	if hasPrometheus {
		prom := Prometheus{
			User:            pkg.FromStrPtr(env.PopPtr(CC_METRICS_PROMETHEUS_USER)),
			Password:        pkg.FromStrPtr(env.PopPtr(CC_METRICS_PROMETHEUS_PASSWORD)),
			Path:            pkg.FromStrPtr(env.PopPtr(CC_METRICS_PROMETHEUS_PATH)),
			Port:            pkg.FromIntPtr(env.PopPtr(CC_METRICS_PROMETHEUS_PORT)),
			ResponseTimeout: pkg.FromIntPtr(env.PopPtr(CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT)),
		}
		obj, d = types.ObjectValueFrom(ctx, PrometheusSchema, prom)
		integration.Prometheus = obj
	}

	if hasPgPoolII {
		pgpool := PgPoolII{
			Enabled: pkg.FromBoolPtr(env.PopPtr(CC_ENABLE_PGPOOL)),
		}
		obj, d = types.ObjectValueFrom(ctx, PGPoolIISchema, pgpool)
		integration.PgPoolII = obj
	}

	diags.Append(d...)
	return integration
}
