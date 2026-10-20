package application

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.dev/sdk/app"
)

// CommonEnvAttributes maps the environment variables backing an attribute
// every runtime shares to that attribute.
var CommonEnvAttributes = map[string]string{
	app.APP_FOLDER: "app_folder",

	app.CC_PRE_BUILD_HOOK:     "hooks.pre_build",
	app.CC_POST_BUILD_HOOK:    "hooks.post_build",
	app.CC_PRE_RUN_HOOK:       "hooks.pre_run",
	app.CC_RUN_SUCCEEDED_HOOK: "hooks.run_succeed",
	app.CC_RUN_FAILED_HOOK:    "hooks.run_failed",

	app.CC_REDIRECTIONIO_PROJECT_KEY:           "integrations.redirectionio.project_key",
	app.CC_REDIRECTIONIO_INSTANCE_NAME:         "integrations.redirectionio.instance_name",
	attributes.CC_REDIRECTIONIO_BACKEND_PORT:   "integrations.redirectionio.backend_port",
	app.NEW_RELIC_LICENSE_KEY:                  "integrations.newrelic.license_key",
	app.NEW_RELIC_APP_NAME:                     "integrations.newrelic.app_name",
	app.CC_CLAMAV:                              "integrations.clamav.enabled",
	app.CC_VARNISH_FILE:                        "integrations.varnish.config_file",
	app.CC_VARNISH_STORAGE_SIZE:                "integrations.varnish.storage_size",
	app.CC_METRICS_PROMETHEUS_USER:             "integrations.prometheus.user",
	app.CC_METRICS_PROMETHEUS_PASSWORD:         "integrations.prometheus.password",
	app.CC_METRICS_PROMETHEUS_PATH:             "integrations.prometheus.path",
	app.CC_METRICS_PROMETHEUS_PORT:             "integrations.prometheus.port",
	app.CC_METRICS_PROMETHEUS_RESPONSE_TIMEOUT: "integrations.prometheus.response_timeout",
	app.CC_ENABLE_PGPOOL:                       "integrations.pgpoolii.enabled",
}

// ReservedEnvironment rejects, in the environment attribute, the variables
// an attribute of the resource manages: set through environment, they are
// moved to the attribute on the next read and the plan never converges.
// runtimeAttributes maps the runtime's own variables to their attribute, on
// top of CommonEnvAttributes.
func ReservedEnvironment(runtimeAttributes map[string]string) resource.ConfigValidator {
	reserved := maps.Clone(CommonEnvAttributes)
	maps.Copy(reserved, runtimeAttributes)
	return reservedEnvironment{reserved: reserved}
}

type reservedEnvironment struct {
	reserved map[string]string
}

func (v reservedEnvironment) Description(ctx context.Context) string {
	return "environment must not hold a variable managed by an attribute"
}

func (v reservedEnvironment) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v reservedEnvironment) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	var env types.Map
	res.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("environment"), &env)...)
	if res.Diagnostics.HasError() || env.IsNull() || env.IsUnknown() {
		return
	}

	keys := slices.Sorted(maps.Keys(env.Elements()))
	for _, key := range keys {
		attribute, ok := v.reserved[key]
		if !ok {
			continue
		}
		res.Diagnostics.AddAttributeError(
			path.Root("environment").AtMapKey(key),
			"Environment variable managed by an attribute",
			fmt.Sprintf("%s is managed by the %q attribute: remove it from environment and set %s instead.", key, attribute, attribute),
		)
	}
}
