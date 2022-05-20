package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// Convert a native string into a tfsdk one
func fromStr(str string) types.String {
	return types.String{Value: str}
}

// Convert a native int into a tfsdk one
func fromI(i int) types.Int64 {
	return types.Int64{Value: int64(i)}
}

// Convert a native int into a tfsdk one
func fromI64(i int64) types.Int64 {
	return types.Int64{Value: i}
}

// Convert a native int into a tfsdk one
func fromB(b bool) types.Bool {
	return types.Bool{Value: b}
}
