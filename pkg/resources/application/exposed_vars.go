package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

var NullExposedEnv = basetypes.NewMapNull(types.StringType)

func SyncExposedVariables(ctx context.Context, p provider.Provider, applicationID string, env types.Map, diags *diag.Diagnostics) {
	m := map[string]string{}

	if !env.IsUnknown() && !env.IsNull() {
		diags.Append(env.ElementsAs(ctx, &m, false)...)
	}

	envRes := tmp.UpdateExposedEnv(ctx, p.Client(), p.Organization(), applicationID, m)
	if envRes.HasError() {
		diags.AddError("failed to update exposed configuration", envRes.Error().Error())
		return
	}
}

// ReadExposedVariables reads exposed environment variables from API.
// stateValue is used to preserve null vs empty semantics.
func ReadExposedVariables(ctx context.Context, p provider.Provider, applicationID string, stateValue types.Map, diags *diag.Diagnostics) types.Map {
	envRes := tmp.GetExposedEnv(ctx, p.Client(), p.Organization(), applicationID)
	if envRes.HasError() {
		diags.AddError("failed to list exposed configuration", envRes.Error().Error())
		return NullExposedEnv
	}

	env := *envRes.Payload()

	if len(env) == 0 {
		if stateValue.IsNull() {
			return NullExposedEnv
		}
		return types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	m, d := types.MapValueFrom(ctx, types.StringType, env)
	diags.Append(d...)

	return m
}
