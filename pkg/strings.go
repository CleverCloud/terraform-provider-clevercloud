package pkg

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"golang.org/x/exp/constraints"
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
func FromI[I constraints.Integer](i I) types.Int64 {
	return types.Int64Value(int64(i))
}

func FromISO8601(d string, diags *diag.Diagnostics) types.Int64 {
	d = strings.SplitN(d, "[", 2)[0]

	t, err := time.Parse(time.RFC3339, d)
	if err != nil {
		diags.AddError("failed to parse ISO8601 date", "expect: RFC3339, got: "+d)
		return types.Int64Null()
	}
	return types.Int64Value(t.Unix())
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

func SetTo[T any](ctx context.Context, items types.Set, diags *diag.Diagnostics) []T {
	var r []T
	diags.Append(items.ElementsAs(ctx, &r, true)...)
	return r
}

func SetToStringSlice(ctx context.Context, items types.Set, diags *diag.Diagnostics) []string {
	return SetTo[string](ctx, items, diags)
}
