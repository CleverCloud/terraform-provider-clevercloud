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
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/retry"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/applications"
)

// CreateReq represents the request structure for creating an application
type CreateReq struct {
	Client       *client.Client
	Organization string
	Application  models.WannabeApplication
	Environment  map[string]string
	VHosts       []string
	Deployment   *Deployment
}

// CreateRes represents the response from creating an application
type CreateRes struct {
	Application models.ApplicationView
	// TargetCommit is the commit resolved by the git deployment on Update
	// (pushed or already running), empty when no deployment is configured
	TargetCommit string
}

// Deployment contains git deployment configuration
type Deployment struct {
	CleverGitAuth      *http.BasicAuth
	Repository         string
	Commit             *string
	Username, Password *string
}

func (r *CreateRes) GetApp() *models.ApplicationView {
	return &r.Application
}

func (r *CreateRes) GetBuildFlavor() types.String {
	return buildFlavor(&r.Application)
}

// buildFlavor is the build flavor of an application, null without a separate build.
func buildFlavor(app *models.ApplicationView) types.String {
	if !app.SeparateBuild {
		return types.StringNull()
	}
	return types.StringValue(pkg.FlavorName(app.BuildFlavor))
}

var githubOAuthService = "github"

// CreateApp handles the low-level API calls for creating an application
func CreateApp(ctx context.Context, req CreateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	res := &CreateRes{}

	req.Application.SeparateBuild = pkg.Ptr(pkg.Deref(req.Application.BuildFlavor) != "")

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

		app := pkg.First(apps, func(app tmp.OAuthApplicationView) bool {
			return pkg.Deref(app.GitURL) == req.Deployment.Repository
		})
		if app == nil {
			diags.AddError("failed to get repository information", "requested repository does not exists or is not visible")
			return res, diags
		}

		// The platform identifies the repository by oauthService and oauthAppId.
		req.Application.OauthService = &githubOAuthService
		req.Application.OauthAppID = &app.ID
	}

	appRes := retry.WithRetry(ctx, func() client.Response[models.ApplicationView] {
		return pkg.SDK(req.Client).V2().Organisations().ID(req.Organization).Applications().Createorganisationapplication(ctx, &req.Application)
	}, fmt.Sprintf("CreateApp(%s)", pkg.Deref(req.Application.Name)))
	if appRes.HasError() {
		diags.AddError("failed to create application", appRes.Error().Error())
		tflog.Error(ctx, "failed to create app", map[string]any{"error": appRes.Error().Error(), "payload": fmt.Sprintf("%+v", req.Application)})
		return nil, diags
	}
	res.Application = *appRes.Payload()

	// Environment
	envRes := UpdateAppEnv(ctx, req.Client, req.Organization, res.Application.ID, req.Environment, &diags)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
	}

	// VHosts
	SyncVHostsOnCreate(ctx, req.Client, req.Organization, req.VHosts, &diags, res.Application.ID)

	// This is dirty, but we need a refresh
	vhostsRes := pkg.SDK(req.Client).V2().Organisations().ID(req.Organization).Applications().Appid(res.Application.ID).Vhosts().Listorganisationapplicationvirtualhosts(ctx)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return res, diags
	}
	res.Application.Vhosts = *vhostsRes.Payload()

	return res, diags
}

// Create centralizes the common Create logic for all application runtimes
func Create[T RuntimePlan](ctx context.Context, resource RuntimeResource, plan T) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Lookup instance by variant slug
	instance := LookupInstanceByVariantSlug(ctx, resource.Client(), nil, resource.GetVariantSlug(), &diags)

	// Extract vhosts and environment from plan
	vhosts := plan.VHostsAsStrings(ctx, &diags)
	environment := plan.ToEnv(ctx, &diags)
	if diags.HasError() {
		return diags
	}

	// Get runtime pointer to access common fields
	runtime := plan.GetRuntimePtr()

	// Build CreateReq
	createReq := CreateReq{
		Client:       resource.Client(),
		Organization: resource.Organization(),
		Application:  wannabeApplication(runtime, instance),
		Environment:  environment,
		VHosts:       vhosts,
		Deployment:   plan.ToDeployment(resource.GitAuth()),
	}

	// Call common Create function
	createRes, createDiags := CreateApp(ctx, createReq)
	diags.Append(createDiags...)

	// Map response even if there were errors (app might be created)
	if createRes != nil {
		runtime.ID = pkg.FromStr(createRes.Application.ID)
		runtime.FromApp(ctx, createRes, &diags)

		// TCP redirection
		runtime.Redirection = SyncTCPRedirection(ctx, resource.Client(), resource.Organization(), createRes.Application.ID, runtime.Redirection, nil, &diags)

		runtime.Tags = SyncTags(ctx, resource, createRes.Application.ID, runtime.Tags, resources.NullTags, &diags)
	}

	return diags
}

// this version of UpdateAppEnv first try with Env var validation reports warning
// and perform the update without validation in order to not break anything
// diag parameter only reports warnings, no error
//
//		{
//		  "id":557,
//		  "message":"Application environment is invalid",
//		  "type":"error",
//		  "fields":{
//		    "CC_TROUBLESHOOT":"must be one of: [troubleshoot, true, 1, yes, enable, enabled, on] (= true) or [off, disable, no, false, disabled, 0] (= false)"
//	   }
//		}
func UpdateAppEnv(ctx context.Context, cc *client.Client, organisationID, applicationID string, envs map[string]string, diags *diag.Diagnostics) client.Response[models.ApplicationView] {
	res := pkg.SDK(cc).V2().Organisations().ID(organisationID).Applications().Appid(applicationID).Env().Updateorganisationapplicationenvironmentvariables(ctx, envs, applications.WithCheck(true))
	if !res.HasError() {
		return res
	}

	// report errors
	if apiErr, ok := res.Error().(*client.APIError); res.StatusCode() == 400 && ok {
		for key, value := range apiErr.Context {
			if key == "type" {
				continue
			}

			diags.AddWarning(fmt.Sprintf("Invalid value for %s", key), fmt.Sprintf("%v", value))
		}
	}

	return pkg.SDK(cc).V2().Organisations().ID(organisationID).Applications().Appid(applicationID).Env().Updateorganisationapplicationenvironmentvariables(ctx, envs)
}

// wannabeApplication is the payload creating or updating an application. Every field
// is set, as the hand-written payload this replaces always sent them all.
func wannabeApplication(runtime *Runtime, instance *tmp.AvailableInstanceView) models.WannabeApplication {
	return models.WannabeApplication{
		Name:           pkg.Ptr(runtime.Name.ValueString()),
		Deploy:         pkg.Ptr("git"),
		Description:    pkg.Ptr(runtime.Description.ValueString()),
		BuildFlavor:    pkg.Ptr(runtime.BuildFlavor.ValueString()),
		MinFlavor:      pkg.Ptr(runtime.SmallestFlavor.ValueString()),
		MaxFlavor:      pkg.Ptr(runtime.BiggestFlavor.ValueString()),
		MinInstances:   pkg.Ptr(models.InstanceCount(runtime.MinInstanceCount.ValueInt64())),
		MaxInstances:   pkg.Ptr(models.InstanceCount(runtime.MaxInstanceCount.ValueInt64())),
		StickySessions: pkg.Ptr(runtime.StickySessions.ValueBool()),
		ForceHTTPS:     pkg.Ptr(FromForceHTTPS(runtime.RedirectHTTPS.ValueBool())),
		Zone:           pkg.Ptr(runtime.Region.ValueString()),
		CancelOnPush:   pkg.Ptr(false),
		SeparateBuild:  pkg.Ptr(false),
		// Create and Update stop on the diagnostics of the lookup, so instance is set here.
		InstanceType:    pkg.Ptr(instance.Type),
		InstanceVariant: pkg.Ptr(instance.Variant.ID),
		InstanceVersion: pkg.Ptr(instance.Version),
	}
}
