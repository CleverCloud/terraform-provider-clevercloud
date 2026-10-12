package resources

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.dev/client"
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

func TestReadTags(t *testing.T) {
	ctx := context.Background()

	t.Run("unmanaged tags are not fetched", func(t *testing.T) {
		diags := diag.Diagnostics{}
		got := ReadTags(ctx, NullTags, func() client.Response[[]string] {
			t.Fatal("fetch must not be called")
			return nil
		}, &diags)
		if diags.HasError() || !got.IsNull() {
			t.Fatalf("got %v (%v), want null", got, diags)
		}
	})

	t.Run("api error keeps the state value", func(t *testing.T) {
		diags := diag.Diagnostics{}
		state := mustSet(t, "a")
		got := ReadTags(ctx, state, func() client.Response[[]string] {
			return fakeResponse{err: errors.New("boom")}
		}, &diags)
		if !diags.HasError() {
			t.Fatal("expected an error")
		}
		if !got.Equal(state) {
			t.Fatalf("got %v, want %v", got, state)
		}
	})

	t.Run("api wins once managed", func(t *testing.T) {
		diags := diag.Diagnostics{}
		got := ReadTags(ctx, mustSet(t, "a"), func() client.Response[[]string] {
			return fakeResponse{payload: []string{"b"}}
		}, &diags)
		if diags.HasError() || !got.Equal(mustSet(t, "b")) {
			t.Fatalf("got %v (%v), want [b]", got, diags)
		}
	})
}

func TestSyncTags(t *testing.T) {
	ctx := context.Background()
	current := mustSet(t, "old")
	planned := mustSet(t, "new", "env:prod")

	t.Run("null plan sends nothing", func(t *testing.T) {
		diags := diag.Diagnostics{}
		got := SyncTags(ctx, NullTags, current, func([]string) client.Response[[]string] {
			t.Fatal("replace must not be called")
			return nil
		}, &diags)
		if diags.HasError() || !got.IsNull() {
			t.Fatalf("got %v (%v), want null", got, diags)
		}
	})

	t.Run("success stores the plan", func(t *testing.T) {
		diags := diag.Diagnostics{}
		var sent []string
		got := SyncTags(ctx, planned, current, func(tags []string) client.Response[[]string] {
			sent = tags
			return fakeResponse{payload: tags}
		}, &diags)
		if diags.HasError() || !got.Equal(planned) {
			t.Fatalf("got %v (%v), want %v", got, diags, planned)
		}
		slices.Sort(sent)
		if !slices.Equal(sent, []string{"env:prod", "new"}) {
			t.Fatalf("sent %v", sent)
		}
	})

	t.Run("failure keeps the current value", func(t *testing.T) {
		diags := diag.Diagnostics{}
		got := SyncTags(ctx, planned, current, func([]string) client.Response[[]string] {
			return fakeResponse{err: errors.New("boom")}
		}, &diags)
		if !diags.HasError() {
			t.Fatal("expected an error")
		}
		if !got.Equal(current) {
			t.Fatalf("got %v, want %v", got, current)
		}
	})
}

// fakeResponse is a canned API response, the client does not export one
type fakeResponse struct {
	payload []string
	err     error
}

func (r fakeResponse) Error() error                         { return r.err }
func (r fakeResponse) HasError() bool                       { return r.err != nil }
func (r fakeResponse) StatusCode() int                      { return 200 }
func (r fakeResponse) IsNotFoundError() bool                { return false }
func (r fakeResponse) SozuID() string                       { return "" }
func (r fakeResponse) Equal(client.Response[[]string]) bool { return false }
func (r fakeResponse) Payload() *[]string                   { return &r.payload }

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
