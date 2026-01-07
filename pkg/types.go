package pkg

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

func IfSetObject[T any](ctx context.Context, diags *diag.Diagnostics, o types.Object, fn func(T)) {
	if o.IsNull() || o.IsUnknown() {
		return
	}

	var t T
	d := o.As(ctx, &t, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if d.HasError() {
		return
	}

	fn(t)
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
