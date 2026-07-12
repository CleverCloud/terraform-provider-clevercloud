package pkg

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.dev/sdk/models"
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

func FromStrMaxLen(str models.StringMaxLength128) types.String {
	return FromStr(string(str))
}

// Convert a native int64 into a tfsdk one
func FromI[I constraints.Integer](i I) types.Int64 {
	return types.Int64Value(int64(i))
}

// Convert a native bool into a tfsdk one
func FromBool(b bool) types.Bool {
	return types.BoolValue(b)
}

// Convert a native float64 into a tfsdk one
func FromFloat64(f float64) types.Float64 {
	return types.Float64Value(f)
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
	// Handle null or unknown Set gracefully
	if items.IsNull() || items.IsUnknown() {
		return []string{}
	}
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

// SetBoolIf sets target to true only if *string equals expected value.
// If condition doesn't match, the target is left unchanged (preserving plan/state value).
// Use this for "magic value" bools like CC_PHP_DEV_DEPENDENCIES="install".
func SetBoolIf(target *types.Bool, s *string, expected string) {
	if s != nil && *s == expected {
		*target = types.BoolValue(true)
	}
}

// MergeTags returns the sorted union of the provider-level default tags and the
// resource-level tags — the effective set applied to a taggable resource.
func MergeTags(defaultTags, resourceTags []string) []string {
	set := map[string]bool{}
	for _, tag := range defaultTags {
		set[tag] = true
	}
	for _, tag := range resourceTags {
		set[tag] = true
	}

	out := make([]string, 0, len(set))
	for tag := range set {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

// ComputeTagsAll computes the effective tag set (provider defaultTags merged with the
// resource-level tags) as a types.Set. It is meant to be called from a resource's ModifyPlan
// to populate the computed `tags_all` attribute — recomputing it at plan time is what makes a
// provider-level default_tags change propagate to existing resources.
//
// If the resource tags are unknown, the result is unknown too.
func ComputeTagsAll(ctx context.Context, defaultTags []string, resourceTags types.Set, diags *diag.Diagnostics) types.Set {
	if resourceTags.IsUnknown() {
		return basetypes.NewSetUnknown(types.StringType)
	}
	return FromSetString(MergeTags(defaultTags, SetToStringSlice(ctx, resourceTags, diags)), diags)
}

// SplitTags separates the full tag set returned by the API into the resource-level `tags`
// (API tags minus the provider default_tags) and the effective `tags_all` (all API tags).
//
// It preserves a null `tags` when the user never declared any (stateTags is null) and there
// are no resource-level tags on the API, to avoid a spurious diff on the Optional attribute.
func SplitTags(apiTags, defaultTags []string, stateTags types.Set, diags *diag.Diagnostics) (tags, tagsAll types.Set) {
	defaults := map[string]bool{}
	for _, tag := range defaultTags {
		defaults[tag] = true
	}

	resourceTags := []string{}
	for _, tag := range apiTags {
		if !defaults[tag] {
			resourceTags = append(resourceTags, tag)
		}
	}

	tagsAll = FromSetString(apiTags, diags)
	if stateTags.IsNull() && len(resourceTags) == 0 {
		return stateTags, tagsAll
	}
	return FromSetString(resourceTags, diags), tagsAll
}

// FromSetSplit splits *string by separator and returns types.Set.
// Returns null set if pointer is nil or string is empty.
func FromSetSplit(s *string, sep string, diags *diag.Diagnostics) types.Set {
	if s == nil || *s == "" {
		return types.SetNull(types.StringType)
	}
	return FromSetString(strings.Split(*s, sep), diags)
}
