package resources

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
)

// tagsValidator rejects tags the API refuses: empty tags, and key/value tags
// with an empty key or an empty value.
var tagsValidator = pkg.NewSetValidator(
	"tags must be non-empty, and 'key:value' tags need both a key and a value",
	func(ctx context.Context, req validator.SetRequest, res *validator.SetResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		for _, tag := range pkg.SetToStringSlice(ctx, req.ConfigValue, &res.Diagnostics) {
			if reason := invalidTagReason(tag); reason != "" {
				res.Diagnostics.AddAttributeError(req.Path, "invalid tag", fmt.Sprintf("tag '%s': %s", tag, reason))
			}
		}
	},
)

func invalidTagReason(tag string) string {
	if strings.TrimSpace(tag) == "" {
		return "a tag cannot be empty"
	}
	if key, value, found := strings.Cut(tag, ":"); found {
		if key == "" {
			return "the key of a 'key:value' tag cannot be empty"
		}
		if value == "" {
			return "the value of a 'key:value' tag cannot be empty"
		}
	}
	return ""
}

// TagsAttribute is the shared schema for the "tags" attribute of applications
// and add-ons. When it is absent from the configuration, tags set outside
// Terraform are left untouched.
var TagsAttribute = schema.SetAttribute{
	ElementType:         types.StringType,
	Optional:            true,
	MarkdownDescription: "Tags of the resource, either plain (`prod`) or key/value (`env:prod`). When unset, tags managed outside Terraform are left untouched and never show up in the state, imports included",
	Validators:          []validator.Set{tagsValidator},
}

// TagsFromAPI maps the tags returned by the API to the state value.
// A null state stays null whatever the API holds: tags are not managed by
// this configuration, so they must never drift. Once managed, the API list is
// the truth, an empty list included.
func TagsFromAPI(ctx context.Context, apiTags []string, stateValue types.Set, diags *diag.Diagnostics) types.Set {
	if stateValue.IsNull() || stateValue.IsUnknown() {
		return types.SetNull(types.StringType)
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
// see TagsFromAPI for the null/empty semantics.
func ReadTags(ctx context.Context, stateValue types.Set, fetch func() client.Response[[]string], diags *diag.Diagnostics) types.Set {
	res := fetch()
	if res.HasError() {
		diags.AddError("failed to get tags", res.Error().Error())
		return types.SetNull(types.StringType)
	}

	return TagsFromAPI(ctx, *res.Payload(), stateValue, diags)
}

// SyncTags replaces the tags of a resource with the planned ones.
// A null or unknown plan means the tags are not managed by Terraform, nothing
// is sent.
func SyncTags(ctx context.Context, planned types.Set, replace func(tags []string) client.Response[[]string], diags *diag.Diagnostics) {
	if planned.IsNull() || planned.IsUnknown() {
		return
	}

	tags := pkg.SetToStringSlice(ctx, planned, diags)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "SYNC TAGS", map[string]any{"tags": tags})

	res := replace(tags)
	if res.HasError() {
		diags.AddError("failed to set tags", res.Error().Error())
	}
}

// NullTags is the tags value of a state that predates the attribute
var NullTags = types.SetNull(types.StringType)
