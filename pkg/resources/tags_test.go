package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTagsFromAPI(t *testing.T) {
	ctx := context.Background()
	empty, _ := types.SetValueFrom(ctx, types.StringType, []string{})

	cases := []struct {
		name  string
		api   []string
		state types.Set
		want  types.Set
	}{
		{"api empty, state null -> null", nil, types.SetNull(types.StringType), types.SetNull(types.StringType)},
		{"api empty, state unknown -> null", nil, types.SetUnknown(types.StringType), types.SetNull(types.StringType)},
		{"api empty, state empty -> empty", []string{}, empty, empty},
		{"api empty, state set -> empty", nil, mustSet(t, "a"), empty},
		{"api set, state null -> null (unmanaged)", []string{"a", "env:prod"}, types.SetNull(types.StringType), types.SetNull(types.StringType)},
		{"api set, state unknown -> null", []string{"a"}, types.SetUnknown(types.StringType), types.SetNull(types.StringType)},
		{"api set, state empty -> set", []string{"a", "env:prod"}, empty, mustSet(t, "a", "env:prod")},
		{"api set, state set -> api wins", []string{"b"}, mustSet(t, "a"), mustSet(t, "b")},
		{"api duplicates -> deduped", []string{"b", "a", "b"}, mustSet(t, "a"), mustSet(t, "a", "b")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := diag.Diagnostics{}
			got := TagsFromAPI(ctx, tc.api, tc.state, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTagsValidator(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		tags    []string
		wantErr bool
	}{
		{[]string{"simple", "key:value", "a:b:c"}, false},
		{[]string{""}, true},
		{[]string{":value"}, true},
		{[]string{"key:"}, true},
		{[]string{" "}, true},
	}

	for _, tc := range cases {
		res := &validator.SetResponse{}
		tagsValidator.ValidateSet(ctx, validator.SetRequest{
			Path:        path.Root("tags"),
			ConfigValue: mustSet(t, tc.tags...),
		}, res)
		if res.Diagnostics.HasError() != tc.wantErr {
			t.Errorf("tags %q: hasError=%v, want %v (%v)", tc.tags, res.Diagnostics.HasError(), tc.wantErr, res.Diagnostics)
		}
	}

	// null and unknown must pass
	for _, v := range []types.Set{types.SetNull(types.StringType), types.SetUnknown(types.StringType)} {
		res := &validator.SetResponse{}
		tagsValidator.ValidateSet(ctx, validator.SetRequest{Path: path.Root("tags"), ConfigValue: v}, res)
		if res.Diagnostics.HasError() {
			t.Errorf("null/unknown set must not error: %v", res.Diagnostics)
		}
	}
}

func mustSet(t *testing.T, items ...string) types.Set {
	t.Helper()
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		values = append(values, types.StringValue(item))
	}
	s, d := types.SetValue(types.StringType, values)
	if d.HasError() {
		t.Fatalf("set: %v", d)
	}
	return s
}
