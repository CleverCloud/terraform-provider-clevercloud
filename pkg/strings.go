package pkg

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Convert a native string into a tfsdk one
// empty string => null tf value
func FromStr(str string) types.String {
	if str == "" {
		return types.StringNull()
	}

	return types.StringValue(str)
}

// Convert a native int64 into a tfsdk one
func FromI(i int64) types.Int64 {
	return types.Int64Value(i)
}

// Convert a native bool into a tfsdk one
func FromBool(b bool) types.Bool {
	return types.BoolValue(b)
}

// Convert a native int64 into a tfsdk one
func FromListString(items []string) types.List {
	if len(items) == 0 {
		return types.ListNull(types.StringType)
	}

	return types.ListValueMust(
		types.StringType,
		Map(items, func(item string) attr.Value {
			return types.StringValue(item)
		}))
}

// Convert a native int64 into a tfsdk one
func FromSetString(items []string, diags *diag.Diagnostics) types.Set {
	s, d := basetypes.NewSetValue(types.StringType, Map(items, func(item string) attr.Value {
		return types.StringValue(item)
	}))
	diags.Append(d...)

	return s
}

func SetToStringSlice(ctx context.Context, items types.Set, diags *diag.Diagnostics) []string {
	var strs []string
	diags.Append(items.ElementsAs(ctx, &strs, true)...)

	return strs
}
