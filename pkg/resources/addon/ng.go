package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.dev/client"
)

func SyncNetworkGroups(
	ctx context.Context,
	cc *client.Client,
	orgID, addonID string,
	ngSet types.Set,
	diags *diag.Diagnostics,
) {
	ngConfigs := pkg.SetTo[resources.NetworkgroupConfig](ctx, ngSet, diags)

	resources.SyncNetworkGroups(
		ctx,
		cc,
		"ADDON",
		orgID,
		addonID,
		ngConfigs,
		diags)
}
