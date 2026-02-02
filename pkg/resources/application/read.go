package application

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	helpermaps "github.com/miton18/helper/maps"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// ReadAppRes represents the response from reading an application
type ReadAppRes struct {
	App          tmp.AppResponse
	AppIsDeleted bool
	Env          []tmp.Env
}

func (res *ReadAppRes) GetApp() *tmp.AppResponse {
	return &res.App
}

func (res *ReadAppRes) GetBuildFlavor() types.String {
	if !res.App.SeparateBuild {
		return types.StringNull()
	}
	return types.StringValue(res.App.BuildFlavor.Name)
}

func (r ReadAppRes) EnvAsMap() *helpermaps.Map[string, string] {
	return pkg.Reduce(
		r.Env,
		helpermaps.NewMap(map[string]string{}),
		func(acc *helpermaps.Map[string, string], entry tmp.Env) *helpermaps.Map[string, string] {
			acc.Set(entry.Name, entry.Value)
			return acc
		})
}

// ReadApp handles the low-level API calls for reading an application
func ReadApp(ctx context.Context, cc *client.Client, orgId, appId string) (*ReadAppRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	r := &ReadAppRes{}

	appRes := tmp.GetApp(ctx, cc, orgId, appId)
	if appRes.IsNotFoundError() {
		r.AppIsDeleted = true
		return r, diags
	}
	if appRes.HasError() {
		diags.AddError("failed to get app", appRes.Error().Error())
		return r, diags
	}

	r.App = *appRes.Payload()

	envRes := tmp.GetAppEnv(ctx, cc, orgId, appId)
	if envRes.IsNotFoundError() {
		r.AppIsDeleted = true
		return r, diags
	}
	if envRes.HasError() {
		diags.AddError("failed to get app", appRes.Error().Error())
		return r, diags
	}

	r.Env = *envRes.Payload()

	return r, diags
}

// Read centralizes the common Read logic for all application runtimes
// Returns true if the app is deleted, false otherwise
func Read[T RuntimePlan](ctx context.Context, resource RuntimeResource, state T) (appIsDeleted bool, diags diag.Diagnostics) {
	// Get runtime pointer to access ID
	runtime := state.GetRuntimePtr()

	// Call the common ReadApp function
	readRes, readDiags := ReadApp(ctx, resource.Client(), resource.Organization(), runtime.ID.ValueString())
	diags.Append(readDiags...)
	if diags.HasError() {
		return false, diags
	}

	// Check if app was deleted
	if readRes.AppIsDeleted {
		return true, diags
	}

	// mutable env map, which will be subset by next reader
	env := readRes.EnvAsMap()

	// Read network groups
	runtime.Networkgroups = resources.ReadNetworkGroups(ctx, resource, runtime.ID.ValueString(), &diags)

	// Read exposed environment variables
	runtime.ExposedEnvironment = ReadExposedVariables(ctx, resource, runtime.ID.ValueString(), runtime.ExposedEnvironment, &diags)

	// Read linked dependencies (addons)
	runtime.Dependencies = ReadDependencies(ctx, resource.Client(), resource.Organization(), runtime.ID.ValueString(), runtime.Dependencies, &diags)

	m, d := types.MapValueFrom(ctx, types.StringType, readRes.EnvAsMap())
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}
	runtime.Environment = m

	// Map environment variables to runtime-specific fields
	state.FromEnv(ctx, env, &diags)

	runtime.Hooks = attributes.FromEnvHooks(env, runtime.Hooks)

	if env.Size() > 0 {
		nativeEnvMap := maps.Collect(env.All)
		m, d := types.MapValueFrom(ctx, types.StringType, nativeEnvMap)
		diags.Append(d...)
		runtime.Environment = m
	} else {
		if runtime.Environment.IsNull() {
			runtime.Environment = types.MapNull(types.StringType)
		} else {
			runtime.Environment = types.MapValueMust(types.StringType, map[string]attr.Value{})
		}
	}

	// Map API response to state
	runtime.SetFromResponse(readRes, ctx, &diags)

	return false, diags
}
