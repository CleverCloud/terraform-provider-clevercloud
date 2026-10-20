package application

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.dev/sdk/app"
)

// AppFolderEnv maps app_folder, shared by every runtime, to APP_FOLDER.
var AppFolderEnv = pkg.NewStringEnv[Runtime](app.APP_FOLDER,
	func(r *Runtime) types.String { return r.AppFolder },
	func(r *Runtime, v types.String) { r.AppFolder = v })

// OnRuntime lets a runtime plan list a mapping of the Runtime it embeds
// among its own.
func OnRuntime[S RuntimePlan](m pkg.EnvVarMapping[*Runtime]) pkg.EnvVarMapping[S] {
	return onRuntime[S]{m}
}

type onRuntime[S RuntimePlan] struct {
	mapping pkg.EnvVarMapping[*Runtime]
}

func (o onRuntime[S]) Name() string               { return o.mapping.Name() }
func (o onRuntime[S]) FromEnv(s S, v string)      { o.mapping.FromEnv(s.GetRuntimePtr(), v) }
func (o onRuntime[S]) ToEnv(s S) (string, string) { return o.mapping.ToEnv(s.GetRuntimePtr()) }

// CommonEnvNames lists the environment variables managed by the hooks and
// integrations blocks every runtime shares.
func CommonEnvNames() []string {
	return slices.Concat(pkg.EnvNames(attributes.HooksEnv...), attributes.IntegrationsEnvNames)
}

// ReservedEnvironment rejects, in the environment attribute, the variables
// of mappings and the shared ones (CommonEnvNames): set through environment,
// they are moved to their attribute on the next read and the plan never
// converges.
func ReservedEnvironment[S any](mappings ...pkg.EnvVarMapping[S]) resource.ConfigValidator {
	return reservedEnvironment{names: slices.Concat(pkg.EnvNames(mappings...), CommonEnvNames())}
}

type reservedEnvironment struct {
	names []string
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

	for _, name := range v.names {
		if _, ok := env.Elements()[name]; !ok {
			continue
		}
		res.Diagnostics.AddAttributeError(
			path.Root("environment").AtMapKey(name),
			"Environment variable managed by an attribute",
			fmt.Sprintf("%s is managed by an attribute of this resource: remove it from environment and set the attribute documented with %s instead.", name, name),
		)
	}
}
