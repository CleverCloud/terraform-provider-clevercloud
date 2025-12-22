package pkg

import (
	"context"
	"strconv"
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

// FromStrPtr converts *string to types.String.
// Returns null if pointer is nil or string is empty.
func FromStrPtr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// FromIntPtr parses *string to types.Int64.
// Returns null if pointer is nil or parsing fails.
func FromIntPtr(s *string) types.Int64 {
	if s == nil {
		return types.Int64Null()
	}
	if i, err := strconv.ParseInt(*s, 10, 64); err == nil {
		return types.Int64Value(i)
	}
	return types.Int64Null()
}

// FromBoolPtr parses *string to types.Bool using strconv.ParseBool.
// Returns null if pointer is nil or parsing fails.
func FromBoolPtr(s *string) types.Bool {
	if s == nil {
		return types.BoolNull()
	}
	if b, err := strconv.ParseBool(*s); err == nil {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

// FromBoolIf returns types.Bool true if *string equals expected value.
// Returns null otherwise.
// WARNING: This overwrites the target. For "magic value" bools where you want
// to preserve the existing value when condition doesn't match, use SetBoolIf instead.
func FromBoolIf(s *string, expected string) types.Bool {
	if s != nil && *s == expected {
		return types.BoolValue(true)
	}
	return types.BoolNull()
}

// SetBoolIf sets target to true only if *string equals expected value.
// If condition doesn't match, the target is left unchanged (preserving plan/state value).
// Use this for "magic value" bools like CC_PHP_DEV_DEPENDENCIES="install".
func SetBoolIf(target *types.Bool, s *string, expected string) {
	if s != nil && *s == expected {
		*target = types.BoolValue(true)
	}
}

// FromSetSplit splits *string by separator and returns types.Set.
// Returns null set if pointer is nil or string is empty.
func FromSetSplit(s *string, sep string, diags *diag.Diagnostics) types.Set {
	if s == nil || *s == "" {
		return types.SetNull(types.StringType)
	}
	return FromSetString(strings.Split(*s, sep), diags)
}
