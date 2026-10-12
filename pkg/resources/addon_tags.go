package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// The add-on tags routes only accept the "addon_xxx" ID, while most resources
// store the provider real ID (postgresql_xxx, ...). Both helpers resolve the
// ID first, a no-op when it already is an add-on ID.

// ReadAddonTags reads the tags of an add-on, see ReadTags for the null/empty
// semantics
func ReadAddonTags(ctx context.Context, p provider.Provider, id string, stateValue types.Set, diags *diag.Diagnostics) types.Set {
	if stateValue.IsNull() || stateValue.IsUnknown() {
		return NullTags
	}

	addonID, err := tmp.RealIDToAddonID(ctx, p.Client(), p.Organization(), id)
	if err != nil {
		diags.AddError("failed to resolve add-on ID", err.Error())
		return stateValue
	}

	return ReadTags(ctx, stateValue, func() client.Response[[]string] {
		return tmp.GetAddonTags(ctx, p.Client(), p.Organization(), addonID)
	}, diags)
}

// SyncAddonTags replaces the tags of an add-on with the planned ones, see
// SyncTags for the returned value
func SyncAddonTags(ctx context.Context, p provider.Provider, id string, planned, current types.Set, diags *diag.Diagnostics) types.Set {
	if planned.IsNull() || planned.IsUnknown() {
		return NullTags
	}

	addonID, err := tmp.RealIDToAddonID(ctx, p.Client(), p.Organization(), id)
	if err != nil {
		diags.AddError("failed to resolve add-on ID", err.Error())
		return current
	}

	return SyncTags(ctx, planned, current, func(tags []string) client.Response[[]string] {
		return tmp.UpdateAddonTags(ctx, p.Client(), p.Organization(), addonID, tags)
	}, diags)
}
