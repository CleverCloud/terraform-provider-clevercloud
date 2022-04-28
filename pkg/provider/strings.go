package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// Convert a native string into a tfsdk one
func fromStr(str string) types.String {
	return types.String{Value: str}
}

// Convert a native int64 into a tfsdk one
func fromI(i int64) types.Int64 {
	return types.Int64{Value: i}
}
