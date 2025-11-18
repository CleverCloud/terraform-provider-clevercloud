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
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
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
	CleverGitAuth      *http.BasicAuth
	Repository         string
	Commit             *string
	Username, Password *string
}

type CreateRes struct {
	Application tmp.CreatAppResponse
}

var githubOAuthService = "github"

// RuntimeResource interface defines methods required by resources to use GenericCreate
type RuntimeResource interface {
	GetVariantSlug() string
	Client() *client.Client
	Organization() string
	GitAuth() *http.BasicAuth
}

// RuntimePlan interface defines methods required by plan types to use GenericCreate
type RuntimePlan interface {
	VHostsAsStrings(ctx context.Context, diags *diag.Diagnostics) []string
	DependenciesAsString(ctx context.Context, diags *diag.Diagnostics) []string
	ToEnv(ctx context.Context, diags *diag.Diagnostics) map[string]string
	ToDeployment(auth *http.BasicAuth) *Deployment
	GetRuntimePtr() *Runtime
}

func (r *CreateRes) GetBuildFlavor() types.String {
	if !r.Application.SeparateBuild {
		return types.StringNull()
	}

	return types.StringValue(r.Application.BuildFlavor.Name)
}

func Create(ctx context.Context, req CreateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	res := &CreateRes{}

	req.Application.SeparateBuild = req.Application.BuildFlavor != ""

	if req.Deployment != nil &&
		req.Deployment.Commit != nil &&
		strings.HasPrefix(*req.Deployment.Commit, attributes.GITHUB_COMMIT_PREFIX) {

		//grab some informations on the repo
		appsRes := tmp.ListGithubApplications(ctx, req.Client)
		if appsRes.HasError() {
			diags.AddError("failed to list Github application", appsRes.Error().Error())
			return res, diags
		}
		apps := *appsRes.Payload()

		app := pkg.First(apps, func(app tmp.GithubApplication) bool {
			return app.GitURL == req.Deployment.Repository
		})
		if app == nil {
			diags.AddError("failed to get repository information", "requested repository does not exists or is not visible")
			return res, diags
		}

		req.Application.GithubApp = app
		req.Application.OAuthService = &githubOAuthService
		req.Application.OAuthAppID = &app.ID
	}

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

	return res, diags
}

// GenericCreate centralizes the common Create logic for all application runtimes
func GenericCreate[T RuntimePlan](ctx context.Context, resource RuntimeResource, plan T) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Lookup instance by variant slug
	instance := LookupInstanceByVariantSlug(ctx, resource.Client(), nil, resource.GetVariantSlug(), &diags)

	// Extract vhosts, environment, and dependencies from plan
	vhosts := plan.VHostsAsStrings(ctx, &diags)
	environment := plan.ToEnv(ctx, &diags)
	dependencies := plan.DependenciesAsString(ctx, &diags)
	if diags.HasError() {
		return diags
	}

	// Get runtime pointer to access common fields
	runtime := plan.GetRuntimePtr()

	// Build CreateReq
	createReq := CreateReq{
		Client:       resource.Client(),
		Organization: resource.Organization(),
		Application: tmp.CreateAppRequest{
			Name:            runtime.Name.ValueString(),
			Deploy:          "git",
			Description:     runtime.Description.ValueString(),
			InstanceType:    instance.Type,
			InstanceVariant: instance.Variant.ID,
			InstanceVersion: instance.Version,
			BuildFlavor:     runtime.BuildFlavor.ValueString(),
			MinFlavor:       runtime.SmallestFlavor.ValueString(),
			MaxFlavor:       runtime.BiggestFlavor.ValueString(),
			MinInstances:    runtime.MinInstanceCount.ValueInt64(),
			MaxInstances:    runtime.MaxInstanceCount.ValueInt64(),
			StickySessions:  runtime.StickySessions.ValueBool(),
			ForceHttps:      FromForceHTTPS(runtime.RedirectHTTPS.ValueBool()),
			Zone:            runtime.Region.ValueString(),
			CancelOnPush:    false,
		},
		Environment:  environment,
		VHosts:       vhosts,
		Dependencies: dependencies,
	}

	// Call common Create function
	createRes, createDiags := Create(ctx, createReq)
	diags.Append(createDiags...)
	if diags.HasError() {
		return diags
	}

	// Map API response to plan using SetFromCreateResponse
	runtime.SetFromCreateResponse(createRes, ctx, &diags)

	// Sync network groups
	SyncNetworkGroups(
		ctx,
		resource.Client(),
		resource.Organization(),
		createRes.Application.ID,
		runtime.Networkgroups,
		&diags,
	)

	// Git deployment
	GitDeploy(ctx, plan.ToDeployment(resource.GitAuth()), createRes.Application.DeployURL, &diags)

	return diags
}

func Read(ctx context.Context, cc *client.Client, orgId, appId string) (*ReadAppRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	r := &ReadAppRes{}

	appRes := tmp.GetApp(ctx, cc, orgId, appId)
	if appRes.IsNotFoundError() {
		r.AppIsDeleted = true
		return r, diags
	}
	if appRes.HasError() {
		diags.AddError("failed to get app", appRes.Error().Error())
		return r, diags
	}

	r.App = *appRes.Payload()

	envRes := tmp.GetAppEnv(ctx, cc, orgId, appId)
	if envRes.IsNotFoundError() {
		r.AppIsDeleted = true
		return r, diags
	}
	if envRes.HasError() {
		diags.AddError("failed to get app", appRes.Error().Error())
		return r, diags
	}

	r.Env = *envRes.Payload()

	return r, diags
}

func Update(ctx context.Context, req UpdateReq) (*CreateRes, diag.Diagnostics) {
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
	gitDeployed := false
	if req.Deployment != nil {
		GitDeploy(ctx, req.Deployment, res.Application.DeployURL, &diags)
		gitDeployed = true
	}

	// trigger restart of the app if needed (when env change)
	// BUT only if we didn't already trigger a Git deployment (which deploys the new code + env)
	// error id 4014 = cannot redeploy an application which has never been deployed yet (did you git push?)
	if req.TriggerRestart && !gitDeployed {
		restartRes := tmp.RestartApp(ctx, req.Client, req.Organization, res.Application.ID)
		if restartRes.HasError() {
			if apiErr, ok := restartRes.Error().(*client.APIError); !ok || apiErr.Code != "4014" {
				diags.AddError("failed to restart app", restartRes.Error().Error())
				return nil, diags
			}
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

	// Delete vhosts that need to be removed (including default vhost if user specified empty list)
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

type ReadAppRes struct {
	App          tmp.CreatAppResponse
	AppIsDeleted bool
	Env          []tmp.Env
}

func (res *ReadAppRes) GetBuildFlavor() types.String {
	if !res.App.SeparateBuild {
		return types.StringNull()
	}
	return types.StringValue(res.App.BuildFlavor.Name)
}

func (r ReadAppRes) EnvAsMap() map[string]string {
	return pkg.Reduce(
		r.Env,
		map[string]string{},
		func(acc map[string]string, entry tmp.Env) map[string]string {
			acc[entry.Name] = entry.Value
			return acc
		})
}
