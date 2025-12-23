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

// CreateReq represents the request structure for creating an application
type CreateReq struct {
	Client       *client.Client
	Organization string
	Application  tmp.CreateAppRequest
	Environment  map[string]string
	VHosts       []string
	Deployment   *Deployment
	Dependencies []string
}

// CreateRes represents the response from creating an application
type CreateRes struct {
	Application tmp.AppResponse
}

// Deployment contains git deployment configuration
type Deployment struct {
	CleverGitAuth      *http.BasicAuth
	Repository         string
	Commit             *string
	Username, Password *string
}

func (r *CreateRes) GetApp() *tmp.AppResponse {
	return &r.Application
}

func (r *CreateRes) GetBuildFlavor() types.String {
	if !r.Application.SeparateBuild {
		return types.StringNull()
	}

	return types.StringValue(r.Application.BuildFlavor.Name)
}

var githubOAuthService = "github"

// CreateApp handles the low-level API calls for creating an application
func CreateApp(ctx context.Context, req CreateReq) (*CreateRes, diag.Diagnostics) {
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
		return res, diags
	}
	res.Application.Vhosts = *vhostsRes.Payload()

	// Dependencies - split into apps and addons as they use different API endpoints
	appDeps, addonDeps := splitDependencies(req.Dependencies)
	tflog.Debug(ctx, "[create] dependencies to link", map[string]any{
		"raw":    req.Dependencies,
		"apps":   appDeps,
		"addons": addonDeps,
	})

	// Link app dependencies using /dependencies endpoint
	for _, appID := range appDeps {
		depRes := tmp.AddAppDependency(ctx, req.Client, req.Organization, res.Application.ID, appID)
		if depRes.HasError() {
			tflog.Error(ctx, "ERROR linking app: "+appID, map[string]any{"err": depRes.Error().Error()})
			diags.AddError("failed to add app dependency", depRes.Error().Error())
		}
	}

	// Link addon dependencies using /addons endpoint
	// Check avoids unnecessary API call to RealIDsToAddonIDs when no addons
	if len(addonDeps) > 0 {
		addonIDs, err := tmp.RealIDsToAddonIDs(ctx, req.Client, req.Organization, addonDeps...)
		if err != nil {
			diags.AddError("failed to get dependencies add-on IDs", err.Error())
			return res, diags
		}
		for _, addonID := range addonIDs {
			depRes := tmp.AddAppLinkedAddons(ctx, req.Client, req.Organization, res.Application.ID, addonID)
			if depRes.HasError() {
				tflog.Error(ctx, "ERROR linking addon: "+addonID, map[string]any{"err": depRes.Error().Error()})
				diags.AddError("failed to add addon dependency", depRes.Error().Error())
			}
		}
	}

	return res, diags
}

// Create centralizes the common Create logic for all application runtimes
func Create[T RuntimePlan](ctx context.Context, resource RuntimeResource, plan T) diag.Diagnostics {
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
		Deployment:   plan.ToDeployment(resource.GitAuth()),
	}

	// Call common Create function
	createRes, createDiags := CreateApp(ctx, createReq)
	diags.Append(createDiags...)

	// Map response even if there were errors (app might be created)
	if createRes != nil {
		runtime.ID = pkg.FromStr(createRes.Application.ID)
		runtime.SetFromResponse(createRes, ctx, &diags)
	}

	return diags
}
