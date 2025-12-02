package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// SyncVHostsOnCreate synchronizes VHosts during resource creation
func SyncVHostsOnCreate(ctx context.Context, client *client.Client, organization string, reqVhosts []string, diags *diag.Diagnostics, applicationID string) {
	// If reqVhosts is nil (not specified), keep default vhosts from API
	// If reqVhosts is an empty slice (explicitly set to []), remove all vhosts
	if reqVhosts == nil {
		return
	}

	// Get current vhosts from remote
	vhostsRes := tmp.GetAppVhosts(ctx, client, organization, applicationID)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return
	}
	remoteVHosts := *vhostsRes.Payload()

	vhostsToAdd := pkg.Diff(reqVhosts, remoteVHosts.AsString())
	vhostsToRemove := pkg.Diff(remoteVHosts.AsString(), reqVhosts)

	tflog.Debug(ctx, "SYNC VHOSTS (CREATE)", map[string]any{
		"planed":   reqVhosts,
		"remote":   remoteVHosts.AsString(),
		"toRemove": vhostsToRemove,
		"toAdd":    vhostsToAdd})

	// Delete vhosts that need to be removed
	for _, vhost := range vhostsToRemove {
		deleteVhostRes := tmp.DeleteAppVHost(ctx, client, organization, applicationID, vhost)
		if deleteVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to remove vhost \"%s\"", vhost), deleteVhostRes.Error().Error())
		}
	}

	// Add new vhosts
	for _, vhost := range vhostsToAdd {
		addVhostRes := tmp.AddAppVHost(ctx, client, organization, applicationID, vhost)
		if addVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to add vhost \"%s\"", vhost), addVhostRes.Error().Error())
		}
	}
}

// SyncVHostsOnUpdate synchronizes VHosts during resource update
func SyncVHostsOnUpdate(ctx context.Context, client *client.Client, organization string, reqVhosts []string, diags *diag.Diagnostics, applicationID string) {
	vhostsRes := tmp.GetAppVhosts(ctx, client, organization, applicationID)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return
	}
	remoteVHosts := *vhostsRes.Payload()

	// What about a creation without vhosts an then an update without vhosts
	/*if len(reqVhosts) == 0 && len(remoteVHosts) == 1 { // expect this

	}*/

	vhostsToAdd := pkg.Diff(reqVhosts, remoteVHosts.AsString())
	vhostsToRemove := pkg.Diff(remoteVHosts.AsString(), reqVhosts)

	tflog.Debug(ctx, "SYNC VHOSTS (UPDATE)", map[string]any{
		"planed":   reqVhosts,
		"remote":   remoteVHosts.AsString(),
		"toRemove": vhostsToRemove,
		"toAdd":    vhostsToAdd})

	// Delete vhosts that need to be removed
	for _, vhost := range vhostsToRemove {
		deleteVhostRes := tmp.DeleteAppVHost(ctx, client, organization, applicationID, vhost)
		if deleteVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to remove vhost \"%s\"", vhost), deleteVhostRes.Error().Error())
		}
	}

	// Add new vhosts
	for _, vhost := range vhostsToAdd {
		addVhostRes := tmp.AddAppVHost(ctx, client, organization, applicationID, vhost)
		if addVhostRes.HasError() {
			diags.AddError(fmt.Sprintf("failed to add vhost \"%s\"", vhost), addVhostRes.Error().Error())
		}
	}
}
