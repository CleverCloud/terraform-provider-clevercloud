package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
)

func SyncNetworkGroups(
	ctx context.Context,
	prov provider.Provider,
	applicationID string,
	ngSet types.Set,
	diags *diag.Diagnostics,
) {
	ngConfigs := pkg.SetTo[resources.NetworkgroupConfig](ctx, ngSet, diags)

	resources.SyncNetworkGroups(
		ctx,
		prov,
		"APPLICATION",
		applicationID,
		ngConfigs,
		diags)
}
