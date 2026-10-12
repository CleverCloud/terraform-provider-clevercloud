package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// ReadTags reads the tags of an application, see resources.TagsFromAPI for
// the null/empty semantics
func ReadTags(ctx context.Context, p provider.Provider, applicationID string, stateValue types.Set, diags *diag.Diagnostics) types.Set {
	return resources.ReadTags(ctx, stateValue, func() client.Response[[]string] {
		return tmp.GetAppTags(ctx, p.Client(), p.Organization(), applicationID)
	}, diags)
}

// SyncTags replaces the tags of an application with the planned ones, a null
// plan leaves them untouched
func SyncTags(ctx context.Context, p provider.Provider, applicationID string, planned types.Set, diags *diag.Diagnostics) {
	resources.SyncTags(ctx, planned, func(tags []string) client.Response[[]string] {
		return tmp.UpdateAppTags(ctx, p.Client(), p.Organization(), applicationID, tags)
	}, diags)
}
