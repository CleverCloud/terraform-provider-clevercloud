package pkg

import (
	"fmt"

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
	fmt.Printf("####### items: %d\n", len(items))
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
func FromSetString(items []string) (types.Set, diag.Diagnostics) {
	return basetypes.NewSetValue(types.StringType, Map(items, func(item string) attr.Value {
		return types.StringValue(item)
	}))
}
