package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// TagsEndpoint describes where a resource kind's tags live on the API.
type TagsEndpoint struct {
	Kind   string // used in error messages and logs
	Get    func(ctx context.Context, cc *client.Client, organisation string, id string) client.Response[[]string]
	Add    func(ctx context.Context, cc *client.Client, organisation string, id string, tag string) client.Response[any]
	Delete func(ctx context.Context, cc *client.Client, organisation string, id string, tag string) client.Response[any]
}

var AddonTags = TagsEndpoint{"addon", tmp.GetAddonTags, tmp.AddAddonTag, tmp.DeleteAddonTag}
var ApplicationTags = TagsEndpoint{"application", tmp.GetAppTags, tmp.AddAppTag, tmp.DeleteAppTag}

// SyncTags reconciles the tags set on a resource with the effective set (provider
// default_tags merged with the resource-level tagsSet): it removes tags that are no
// longer wanted and adds the missing ones. It is order-insensitive and idempotent.
//
// It writes the resource-level tags to stateTarget and the effective merged set to
// tagsAllTarget, so provider tags never leak into the resource's own `tags` attribute
// (which would cause a perpetual diff) while `tags_all` reflects reality.
//
// For AddonTags, id MUST be the addon ID (addon_xxx), not the realId (e.g. cellar_xxx,
// postgresql_xxx) — the tags endpoint does not accept the realId. Use
// tmp.RealIDToAddonID to convert when only the realId is at hand.
func SyncTags(
	ctx context.Context,
	prov provider.Provider,
	ep TagsEndpoint,
	id string,
	tagsSet types.Set,
	stateTarget *types.Set,
	tagsAllTarget *types.Set,
	diags *diag.Diagnostics,
) {
	resourceTags := pkg.SetToStringSlice(ctx, tagsSet, diags)
	if diags.HasError() {
		return
	}
	wanted := pkg.MergeTags(prov.DefaultTags(), resourceTags)

	tagsRes := ep.Get(ctx, prov.Client(), prov.Organization(), id)
	if tagsRes.HasError() {
		diags.AddError("failed to get "+ep.Kind+" tags", tagsRes.Error().Error())
		return
	}

	current := set.New((*tagsRes.Payload())...)
	wantedSet := set.New(wanted...)

	toRemove := current.Difference(wantedSet).Slice()
	toAdd := wantedSet.Difference(current).Slice()
	tflog.Debug(ctx, "syncing "+ep.Kind+" tags", map[string]any{ep.Kind: id, "add": toAdd, "remove": toRemove})

	for _, tag := range toRemove {
		if res := ep.Delete(ctx, prov.Client(), prov.Organization(), id, tag); res.HasError() && !res.IsNotFoundError() {
			diags.AddError("failed to remove "+ep.Kind+" tag", res.Error().Error())
			return
		}
		tflog.Info(ctx, "removed "+ep.Kind+" tag", map[string]any{ep.Kind: id, "tag": tag})
	}
	for _, tag := range toAdd {
		if res := ep.Add(ctx, prov.Client(), prov.Organization(), id, tag); res.HasError() {
			diags.AddError("failed to add "+ep.Kind+" tag", res.Error().Error())
			return
		}
		tflog.Info(ctx, "added "+ep.Kind+" tag", map[string]any{ep.Kind: id, "tag": tag})
	}

	*stateTarget = tagsSet
	*tagsAllTarget = pkg.FromSetString(wanted, diags)
}

// ReadTags fetches the tags currently set on a resource and returns both the
// resource-level tags (the API tags minus the provider default_tags) and the effective
// merged set (`tags_all`, all API tags). It preserves a null state when the user never
// declared tags and there are no resource-level tags on the API, to avoid a spurious
// diff on the Optional attribute.
//
// For AddonTags, id MUST be the addon ID (addon_xxx), not the realId.
func ReadTags(
	ctx context.Context,
	prov provider.Provider,
	ep TagsEndpoint,
	id string,
	stateTags types.Set,
	diags *diag.Diagnostics,
) (tags types.Set, tagsAll types.Set) {
	tagsRes := ep.Get(ctx, prov.Client(), prov.Organization(), id)
	if tagsRes.HasError() {
		diags.AddError("failed to get "+ep.Kind+" tags", tagsRes.Error().Error())
		return stateTags, basetypes.NewSetUnknown(types.StringType)
	}
	return pkg.SplitTags(*tagsRes.Payload(), prov.DefaultTags(), stateTags, diags)
}
