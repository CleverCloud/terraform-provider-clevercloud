package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

// VHostNames returns the FQDN of each vhost, in order.
func VHostNames(vhosts []models.VhostView) []string {
	return pkg.Map(vhosts, func(vhost models.VhostView) string { return vhost.Fqdn })
}

// SyncVHostsOnCreate synchronizes VHosts during resource creation
func SyncVHostsOnCreate(ctx context.Context, client *client.Client, organization string, reqVhosts []string, diags *diag.Diagnostics, applicationID string) {
	// If reqVhosts is nil (not specified), keep default vhosts from API
	// If reqVhosts is an empty slice (explicitly set to []), remove all vhosts
	if reqVhosts == nil {
		return
	}

	// Get current vhosts from remote
	vhostsRes := pkg.SDK(client).V2().Organisations().ID(organization).Applications().Appid(applicationID).Vhosts().Listorganisationapplicationvirtualhosts(ctx)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return
	}
	remoteVHosts := *vhostsRes.Payload()

	vhostsToAdd := pkg.Diff(reqVhosts, VHostNames(remoteVHosts))
	vhostsToRemove := pkg.Diff(VHostNames(remoteVHosts), reqVhosts)

	tflog.Debug(ctx, "SYNC VHOSTS (CREATE)", map[string]any{
		"planed":   reqVhosts,
		"remote":   VHostNames(remoteVHosts),
		"toRemove": vhostsToRemove,
		"toAdd":    vhostsToAdd})

	// Delete vhosts that need to be removed
	for _, vhost := range vhostsToRemove {
		deleteVhostRes := pkg.SDK(client).V2().Organisations().ID(organization).Applications().Appid(applicationID).Vhosts().Domain(vhost).Removeorganisationapplicationvirtualhost(ctx)
		if deleteVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to remove vhost \"%s\"", vhost), deleteVhostRes.Error().Error())
		}
	}

	// Add new vhosts
	for _, vhost := range vhostsToAdd {
		addVhostRes := pkg.SDK(client).V2().Organisations().ID(organization).Applications().Appid(applicationID).Vhosts().Domain(vhost).Addorganisationapplicationvirtualhost(ctx)
		if addVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to add vhost \"%s\"", vhost), addVhostRes.Error().Error())
		}
	}
}

// SyncVHostsOnUpdate synchronizes VHosts during resource update
func SyncVHostsOnUpdate(ctx context.Context, client *client.Client, organization string, reqVhosts []string, diags *diag.Diagnostics, applicationID string) {
	vhostsRes := pkg.SDK(client).V2().Organisations().ID(organization).Applications().Appid(applicationID).Vhosts().Listorganisationapplicationvirtualhosts(ctx)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return
	}
	remoteVHosts := *vhostsRes.Payload()

	// What about a creation without vhosts an then an update without vhosts
	/*if len(reqVhosts) == 0 && len(remoteVHosts) == 1 { // expect this

	}*/

	vhostsToAdd := pkg.Diff(reqVhosts, VHostNames(remoteVHosts))
	vhostsToRemove := pkg.Diff(VHostNames(remoteVHosts), reqVhosts)

	tflog.Debug(ctx, "SYNC VHOSTS (UPDATE)", map[string]any{
		"planed":   reqVhosts,
		"remote":   VHostNames(remoteVHosts),
		"toRemove": vhostsToRemove,
		"toAdd":    vhostsToAdd})

	// Delete vhosts that need to be removed
	for _, vhost := range vhostsToRemove {
		deleteVhostRes := pkg.SDK(client).V2().Organisations().ID(organization).Applications().Appid(applicationID).Vhosts().Domain(vhost).Removeorganisationapplicationvirtualhost(ctx)
		if deleteVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to remove vhost \"%s\"", vhost), deleteVhostRes.Error().Error())
		}
	}

	// Add new vhosts
	for _, vhost := range vhostsToAdd {
		addVhostRes := pkg.SDK(client).V2().Organisations().ID(organization).Applications().Appid(applicationID).Vhosts().Domain(vhost).Addorganisationapplicationvirtualhost(ctx)
		if addVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to add vhost \"%s\"", vhost), addVhostRes.Error().Error())
		}
	}
}

// CleverAppsVHost returns the default `<app>.cleverapps.io/` vhost of an
// application, nil when it has none.
func CleverAppsVHost(vhosts []models.VhostView, appID string) *models.VhostView {
	cleverapps := fmt.Sprintf("%s.cleverapps.io/", strings.ReplaceAll(appID, "_", "-"))
	for _, vhost := range vhosts {
		if vhost.Fqdn == cleverapps {
			return &vhost
		}
	}
	return nil
}
