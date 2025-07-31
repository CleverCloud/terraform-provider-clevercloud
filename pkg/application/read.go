package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ReadAppRes struct {
	App          tmp.CreatAppResponse
	AppIsDeleted bool
	Env          []tmp.Env
	ExposedEnv   map[string]string
}

func (res *ReadAppRes) GetBuildFlavor() types.String {
	if !res.App.SeparateBuild {
		return types.StringNull()
	}
	return types.StringValue(res.App.BuildFlavor.Name)
}

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

	exposedEnvRes := tmp.GetAppExposedEnv(ctx, cc, orgId, appId)
	if exposedEnvRes.HasError() {
		diags.AddWarning("failed to get app exposed environment", exposedEnvRes.Error().Error())
	} else {
		r.ExposedEnv = *exposedEnvRes.Payload()
	}

	return r, diags
}

func (r ReadAppRes) EnvAsMap() map[string]string {
	return pkg.Reduce(
		r.Env,
		map[string]string{},
		func(acc map[string]string, entry tmp.Env) map[string]string {
			acc[entry.Name] = entry.Value
			return acc
		})
}
