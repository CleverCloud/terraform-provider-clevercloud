package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type CreateReq struct {
	Client       *client.Client
	Organization string
	Application  tmp.CreateAppRequest
	Environment  map[string]string
	VHosts       []string
	Deployment   *Deployment
	Dependencies []string
}

type UpdateReq struct {
	ID             string
	Client         *client.Client
	Organization   string
	Application    tmp.UpdateAppReq
	Environment    map[string]string
	VHosts         []string
	Deployment     *Deployment
	Dependencies   []string
	TriggerRestart bool // when env vars change for example
}

type Deployment struct {
	CleverGitAuth  *http.BasicAuth
	Repository     string
	Commit         *string
	User, Password *string
	PrivateSSHKey  *string
}

type CreateRes struct {
	Application tmp.CreatAppResponse
}

func (r *CreateRes) GetBuildFlavor() types.String {
	if !r.Application.SeparateBuild {
		return types.StringNull()
	}

	return types.StringValue(r.Application.BuildFlavor.Name)
}

func CreateApp(ctx context.Context, req CreateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Application
	res := &CreateRes{}

	req.Application.SeparateBuild = req.Application.BuildFlavor != ""

	appRes := tmp.CreateApp(ctx, req.Client, req.Organization, req.Application)
	if appRes.HasError() {
		diags.AddError("failed to create application", appRes.Error().Error())
		tflog.Error(ctx, "failed to create app", map[string]any{"error": appRes.Error().Error(), "payload": fmt.Sprintf("%+v", req.Application)})
		return nil, diags
	}
	res.Application = *appRes.Payload()

	// Environment
	envRes := tmp.UpdateAppEnv(ctx, req.Client, req.Organization, res.Application.ID, req.Environment)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
	}

	// VHosts
	SyncVHostsOnCreate(ctx, req.Client, req.Organization, req.VHosts, &diags, res.Application.ID)

	// This is dirty, but we need a refresh
	vhostsRes := tmp.GetAppVhosts(ctx, req.Client, req.Organization, res.Application.ID)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return nil, diags
	}
	res.Application.Vhosts = *vhostsRes.Payload()

	// Dependencies
	dependenciesWithAddonIDs, err := tmp.RealIDsToAddonIDs(ctx, req.Client, req.Organization, req.Dependencies...)
	if err != nil {
		diags.AddError("failed to get dependencies add-on IDs", err.Error())
		return nil, diags
	}
	tflog.Debug(ctx, "[create] dependencies to link", map[string]any{"dependencies": req.Dependencies, "addonIds": dependenciesWithAddonIDs})
	for _, dependency := range dependenciesWithAddonIDs {
		// TODO: support another apps as dependency

		depRes := tmp.AddAppLinkedAddons(ctx, req.Client, req.Organization, res.Application.ID, dependency)
		if depRes.HasError() {
			tflog.Error(ctx, "ERROR: "+dependency, map[string]any{"err": depRes.Error().Error()})
			diags.AddError("failed to add dependency", depRes.Error().Error())
		}
	}

	// Git Deployment
	if req.Deployment != nil {
		diags.Append(gitDeploy(ctx, *req.Deployment, req.Client, res.Application.DeployURL)...)
	}

	return res, diags
}

func UpdateApp(ctx context.Context, req UpdateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Application
	res := &CreateRes{}

	req.Application.SeparateBuild = req.Application.BuildFlavor != ""

	appRes := tmp.UpdateApp(ctx, req.Client, req.Organization, req.ID, req.Application)
	if appRes.HasError() {
		diags.AddError("failed to update application", appRes.Error().Error())
		tflog.Error(ctx, "failed to update app", map[string]any{"error": appRes.Error().Error(), "payload": fmt.Sprintf("%+v", req.Application)})
		return nil, diags
	}

	res.Application = *appRes.Payload()

	// Environment
	envRes := tmp.UpdateAppEnv(ctx, req.Client, req.Organization, res.Application.ID, req.Environment)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
		return nil, diags
	}

	// VHosts
	SyncVHostsOnUpdate(ctx, req.Client, req.Organization, req.VHosts, &diags, res.Application.ID)

	// This is dirty, but we need a refresh
	vhostsRes := tmp.GetAppVhosts(ctx, req.Client, req.Organization, res.Application.ID)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return nil, diags
	}
	res.Application.Vhosts = *vhostsRes.Payload()

	// Dependencies
	dependenciesWithAddonIDs, err := tmp.RealIDsToAddonIDs(ctx, req.Client, req.Organization, req.Dependencies...)
	if err != nil {
		diags.AddError("failed to get dependencies add-on IDs", err.Error())
		return nil, diags
	}

	tflog.Debug(ctx, "[update] dependencies to link", map[string]any{"dependencies": req.Dependencies, "addonIds": dependenciesWithAddonIDs})

	for _, dependency := range dependenciesWithAddonIDs {
		// TODO: support another apps as dependency

		depRes := tmp.AddAppLinkedAddons(ctx, req.Client, req.Organization, res.Application.ID, dependency)
		if depRes.HasError() {
			tflog.Error(ctx, "ERROR: "+dependency, map[string]any{"err": depRes.Error().Error()})
			diags.AddError("failed to add dependency", depRes.Error().Error())
			return nil, diags
		}
	}
	// TODO: unlink unneeded deps

	// Git Deployment (when commit change)
	if req.Deployment != nil {
		diags.Append(gitDeploy(ctx, *req.Deployment, req.Client, res.Application.DeployURL)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	// trigger restart of the app if needed (when env change)
	// error id 4014 = cannot redeploy an application which has never been deployed yet (did you git push?)
	if req.TriggerRestart {
		restartRes := tmp.RestartApp(ctx, req.Client, req.Organization, res.Application.ID)
		if restartRes.HasError() && !strings.Contains(restartRes.Error().Error(), "4014") {
			diags.AddError("failed to restart app", restartRes.Error().Error())
			return nil, diags
		}
	}

	return res, diags
}

// on clever side, it's an enum
func FromForceHTTPS(force bool) string {
	if force {
		return "ENABLED"
	} else {
		return "DISABLED"
	}
}

func ToForceHTTPS(force string) bool {
	return force == "ENABLED"
}

func SyncVHostsOnCreate(ctx context.Context, client *client.Client, organization string, reqVhosts []string, diags *diag.Diagnostics, applicationID string) {
	if len(reqVhosts) == 0 {
		return // on creation keep default vhosts from API
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
