package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// ReadAppRes represents the response from reading an application
type ReadAppRes struct {
	App          tmp.CreatAppResponse
	AppIsDeleted bool
	Env          []tmp.Env
}

func (res *ReadAppRes) GetBuildFlavor() types.String {
	if !res.App.SeparateBuild {
		return types.StringNull()
	}
	return types.StringValue(res.App.BuildFlavor.Name)
}

func (r ReadAppRes) EnvAsMap() pkg.EnvMap {
	return pkg.Reduce(
		r.Env,
		pkg.EnvMap{},
		func(acc pkg.EnvMap, entry tmp.Env) pkg.EnvMap {
			acc[entry.Name] = entry.Value
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

	// Map API response to state using SetFromReadResponse
	runtime.SetFromReadResponse(readRes, ctx, &diags)

	// Read network groups
	runtime.Networkgroups = resources.ReadNetworkGroups(ctx, resource, runtime.ID.ValueString(), &diags)

	// Read exposed environment variables
	runtime.ExposedEnvironment = ReadExposedVariables(ctx, resource, runtime.ID.ValueString(), &diags)

	// Map environment variables to runtime-specific fields
	state.FromEnv(ctx, readRes.EnvAsMap(), &diags)

	return false, diags
}
