package resources

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
)

// TagsAttribute is the shared schema for the "tags" attribute of applications
// and add-ons. When it is absent from the configuration, tags set outside
// Terraform are left untouched.
var TagsAttribute = schema.SetAttribute{
	ElementType:         types.StringType,
	Optional:            true,
	MarkdownDescription: "Tags of the resource, either plain (`prod`) or key/value (`env:prod`). When unset, tags are not managed by Terraform: the ones set outside of it are left untouched and never show up in the state, imports included. Removing the attribute therefore keeps the current tags, set `tags = []` to remove them all",
}

// TagsFromAPI maps the tags returned by the API to the state value.
// A null state stays null whatever the API holds: tags are not managed by
// this configuration, so they must never drift. Once managed, the API list is
// the truth, an empty list included.
func TagsFromAPI(ctx context.Context, apiTags []string, stateValue types.Set, diags *diag.Diagnostics) types.Set {
	if stateValue.IsNull() || stateValue.IsUnknown() {
		return NullTags
	}

	tags := slices.Clone(apiTags)
	slices.Sort(tags)
	tags = slices.Compact(tags)
	if tags == nil {
		tags = []string{}
	}

	result, d := types.SetValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)
	return result
}

// ReadTags fetches the tags of a resource and maps them to the state value,
// see TagsFromAPI for the null/empty semantics. Unmanaged tags are not
// fetched, and the state value is kept when the API call fails.
func ReadTags(ctx context.Context, stateValue types.Set, fetch func() client.Response[[]string], diags *diag.Diagnostics) types.Set {
	if stateValue.IsNull() || stateValue.IsUnknown() {
		return NullTags
	}

	res := fetch()
	if res.HasError() {
		diags.AddError("failed to get tags", res.Error().Error())
		return stateValue
	}

	return TagsFromAPI(ctx, *res.Payload(), stateValue, diags)
}

// SyncTags replaces the tags of a resource with the planned ones and returns
// the value to store in the state: the planned one on success, current when
// the API call fails. A null or unknown plan means the tags are not managed by
// Terraform, nothing is sent.
func SyncTags(ctx context.Context, planned, current types.Set, replace func(tags []string) client.Response[[]string], diags *diag.Diagnostics) types.Set {
	if planned.IsNull() || planned.IsUnknown() {
		return NullTags
	}

	tags := pkg.SetToStringSlice(ctx, planned, diags)
	if diags.HasError() {
		return current
	}

	tflog.Debug(ctx, "sync tags", map[string]any{"tags": tags})

	res := replace(tags)
	if res.HasError() {
		diags.AddError("failed to set tags", res.Error().Error())
		return current
	}

	return planned
}

// NullTags is the value of unmanaged tags, and of a state that predates the
// attribute
var NullTags = types.SetNull(types.StringType)

// TagRefs turns tags into the payload the SDK's tag replacement routes take.
func TagRefs(tags []string) []*string {
	refs := make([]*string, len(tags))
	for i := range tags {
		refs[i] = &tags[i]
	}
	return refs
}
