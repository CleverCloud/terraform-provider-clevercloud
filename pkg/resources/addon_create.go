package resources

import (
	"context"
	"fmt"

	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/retry"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

// CreateAddon provisions an add-on, retrying while the API answers 503.
func CreateAddon(ctx context.Context, cc *client.Client, organisation string, addon models.WannabeAddonProvision) client.Response[models.AddonView] {
	return retry.WithRetry(ctx, func() client.Response[models.AddonView] {
		return pkg.SDK(cc).V2().Organisations().ID(organisation).Addons().Provisionaddonbyorgaid(ctx, &addon)
	}, fmt.Sprintf("CreateAddon(%s)", addon.Name))
}
