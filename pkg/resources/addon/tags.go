package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
)

// SyncTags replaces the tags of an add-on with the planned ones and stores the
// planned value in the state field. A null plan leaves the add-on untouched.
// id is either the add-on ID or the provider real ID.
func SyncTags(
	ctx context.Context,
	p provider.Provider,
	id string,
	planned types.Set,
	target *types.Set,
	diags *diag.Diagnostics,
) {
	resources.SyncAddonTags(ctx, p, id, planned, diags)
	*target = planned
}
