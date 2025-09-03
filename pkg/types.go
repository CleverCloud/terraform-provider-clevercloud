package pkg

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func IfIsSetStr(v types.String, fn func(s string)) {
	IfIsSet(v, func(v types.String) {
		fn(v.ValueString())
	})
}

func IfIsSetB(v types.Bool, fn func(s bool)) {
	if !v.IsNull() && !v.IsUnknown() {
		fn(v.ValueBool())
	}
}

func IfIsSetI(v types.Int64, fn func(i int64)) {
	if !v.IsNull() && !v.IsUnknown() {
		fn(v.ValueInt64())
	}
}

func AtLeastOneSet(v ...attr.Value) bool {
	return Reduce(v, false, func(acc bool, v attr.Value) bool {
		return acc || !v.IsNull() && !v.IsUnknown()
	})
}

func IfIsSet[T attr.Value](v T, fn func(v T)) {
	if !v.IsNull() && !v.IsUnknown() {
		fn(v)
	}
}
