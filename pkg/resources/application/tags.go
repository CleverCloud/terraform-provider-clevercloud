package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.dev/client"
)

// ReadTags reads the tags of an application, see resources.ReadTags for the
// null/empty semantics
func ReadTags(ctx context.Context, p provider.Provider, applicationID string, stateValue types.Set, diags *diag.Diagnostics) types.Set {
	return resources.ReadTags(ctx, stateValue, func() client.Response[[]string] {
		return pkg.SDK(p.Client()).V2().Organisations().ID(p.Organization()).Applications().Appid(applicationID).Tags().Listorganisationapplicationtags(ctx)
	}, diags)
}

// SyncTags replaces the tags of an application with the planned ones, see
// resources.SyncTags for the returned value
func SyncTags(ctx context.Context, p provider.Provider, applicationID string, planned, current types.Set, diags *diag.Diagnostics) types.Set {
	return resources.SyncTags(ctx, planned, current, func(tags []string) client.Response[[]string] {
		return pkg.SDK(p.Client()).V2().Organisations().ID(p.Organization()).Applications().Appid(applicationID).Tags().Updateorganisationapplicationtags(ctx, resources.TagRefs(tags))
	}, diags)
}
