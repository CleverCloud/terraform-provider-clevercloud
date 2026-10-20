package application

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.dev/sdk/app"
)

// RuntimeEnvVar is an environment variable backing an attribute every
// runtime shares; its value is the variable name.
type RuntimeEnvVar string

// AppFolder backs app_folder.
const AppFolder RuntimeEnvVar = app.APP_FOLDER

func (v RuntimeEnvVar) FromEnv(r *Runtime, value string) { *v.attribute(r) = pkg.FromStr(value) }
func (v RuntimeEnvVar) ToEnv(r *Runtime) (string, string) {
	return string(v), v.attribute(r).ValueString()
}

// attribute returns the runtime attribute v backs.
func (v RuntimeEnvVar) attribute(r *Runtime) *types.String {
	switch v {
	case AppFolder:
		return &r.AppFolder
	}
	panic("no runtime attribute for " + string(v))
}

// OnRuntime lets a runtime plan list a mapping of the Runtime it embeds
// among its own.
func OnRuntime[S RuntimePlan](m pkg.EnvVarMapping[*Runtime]) pkg.EnvVarMapping[S] {
	return onRuntime[S]{m}
}

type onRuntime[S RuntimePlan] struct {
	mapping pkg.EnvVarMapping[*Runtime]
}

func (o onRuntime[S]) FromEnv(s S, v string)      { o.mapping.FromEnv(s.GetRuntimePtr(), v) }
func (o onRuntime[S]) ToEnv(s S) (string, string) { return o.mapping.ToEnv(s.GetRuntimePtr()) }

// CommonEnvNames lists the environment variables managed by the hooks and
// integrations blocks every runtime shares.
func CommonEnvNames() []string {
	return slices.Concat(pkg.EnvNames(&attributes.Hooks{}, attributes.HooksEnv...), attributes.IntegrationsEnvNames)
}

// ReservedEnvironment rejects, in the environment attribute, the variables
// managed by an attribute of the resource: set through environment, they are
// moved to the attribute on the next read and the plan never converges.
// names are the resource's own mapped variables, on top of CommonEnvNames.
func ReservedEnvironment(names ...string) resource.ConfigValidator {
	reserved := map[string]bool{}
	for _, name := range slices.Concat(names, CommonEnvNames()) {
		reserved[name] = true
	}
	return reservedEnvironment{reserved: reserved}
}

type reservedEnvironment struct {
	reserved map[string]bool
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

	for _, key := range slices.Sorted(maps.Keys(env.Elements())) {
		if !v.reserved[key] {
			continue
		}
		res.Diagnostics.AddAttributeError(
			path.Root("environment").AtMapKey(key),
			"Environment variable managed by an attribute",
			fmt.Sprintf("%s is managed by an attribute of this resource: remove it from environment and set the attribute documented with %s instead.", key, key),
		)
	}
}
