package pkg

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func IfIsSet(v types.String, fn func(s string)) {
	if !v.IsNull() && !v.IsUnknown() {
		fn(v.ValueString())
	}
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
