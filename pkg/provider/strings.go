package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Convert a native string into a tfsdk one
// empty string => null tf value
func fromStr(str string) types.String {
	if str == "" {
		return types.StringNull()
	}

	return types.StringValue(str)
}

// Convert a native int64 into a tfsdk one
func fromI(i int64) types.Int64 {
	return types.Int64Value(i)
}

// Convert a native int64 into a tfsdk one
func fromListString(items []string) types.List {
	return types.ListValueMust(
		types.StringType,
		Map(items, func(item string) attr.Value {
			return types.StringValue(item)
		}))
}
