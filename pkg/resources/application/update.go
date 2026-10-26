package application

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/applications"
)

// UpdateReq represents the request structure for updating an application
type UpdateReq struct {
	ID             string
	Client         *client.Client
	Organization   string
	Application    models.WannabeApplication
	Environment    map[string]string
	VHosts         []string
	Deployment     *Deployment
	TriggerRestart bool // when env vars change for example
	TriggerRebuild bool // restart without the build cache, when a build time env var changes
}

// UpdateApp handles the low-level API calls for updating an application
func UpdateApp(ctx context.Context, req UpdateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Application
	res := &CreateRes{}

	req.Application.SeparateBuild = pkg.Ptr(pkg.Deref(req.Application.BuildFlavor) != "")

	appRes := pkg.SDK(req.Client).V2().Organisations().ID(req.Organization).Applications().Appid(req.ID).Updateorganisationapplication(ctx, &req.Application)
	if appRes.HasError() {
		diags.AddError("failed to update application", appRes.Error().Error())
		tflog.Error(ctx, "failed to update app", map[string]any{"error": appRes.Error().Error(), "payload": fmt.Sprintf("%+v", req.Application)})
		return nil, diags
	}

	res.Application = *appRes.Payload()

	// Environment
	envRes := UpdateAppEnv(ctx, req.Client, req.Organization, res.Application.ID, req.Environment, &diags)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
		return res, diags
	}

	// VHosts
	SyncVHostsOnUpdate(ctx, req.Client, req.Organization, req.VHosts, &diags, res.Application.ID)

	// This is dirty, but we need a refresh
	vhostsRes := pkg.SDK(req.Client).V2().Organisations().ID(req.Organization).Applications().Appid(res.Application.ID).Vhosts().Listorganisationapplicationvirtualhosts(ctx)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return res, diags
	}
	res.Application.Vhosts = *vhostsRes.Payload()

	// Git Deployment: the target commit (explicit one or repository HEAD) is
	// compared to the currently deployed commit (CommitID), the push only
	// happens when they differ
	gitDeployed := false
	if req.Deployment != nil {
		res.TargetCommit, gitDeployed = GitDeploy(ctx, req.Deployment, pkg.Deref(res.Application.DeployURL), pkg.Deref(res.Application.CommitID), &diags)
		if diags.HasError() {
			return res, diags
		}
	}

	// trigger restart of the app if needed (when env change)
	// BUT only if we didn't already trigger a Git deployment (which deploys the new code + env)
	// error id 4014 = cannot redeploy an application which has never been deployed yet (did you git push?)
	if req.TriggerRestart && !gitDeployed {
		restartRes := restartApp(ctx, req.Client, req.Organization, res.Application.ID, req.TriggerRebuild)
		if restartRes.HasError() {
			if apiErr, ok := restartRes.Error().(*client.APIError); !ok || apiErr.Code != "4014" {
				diags.AddError("failed to restart app", restartRes.Error().Error())
				return res, diags
			}
		}
	}

	return res, diags
}

// Update centralizes the common Update logic for all application runtimes.
// config is needed on top of plan for the deployment block: `deployment.commit`
// is Optional+Computed, so Terraform copies the previous state value (the last
// deployed hash) into the plan when the attribute is not configured. Deploying
// from the plan would pin that old hash instead of tracking the repository
// HEAD, only the configuration carries the user intent.
func Update[T RuntimePlan](ctx context.Context, resource RuntimeResource, plan, config, state T) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Lookup instance by variant slug
	instance := LookupInstanceByVariantSlug(ctx, resource.Client(), nil, resource.GetVariantSlug(), &diags)
	if diags.HasError() {
		return diags
	}

	// Extract environments from plan and state
	planEnvironment := plan.ToEnv(ctx, &diags)
	if diags.HasError() {
		return diags
	}
	stateEnvironment := state.ToEnv(ctx, &diags)
	if diags.HasError() {
		return diags
	}

	// Extract vhosts from plan
	vhosts := plan.VHostsAsStrings(ctx, &diags)
	if diags.HasError() {
		return diags
	}

	// Get runtime pointer to access common fields
	runtime := plan.GetRuntimePtr()
	stateRuntime := state.GetRuntimePtr()

	triggerRestart := !reflect.DeepEqual(planEnvironment, stateEnvironment)
	if triggerRestart {
		tflog.Debug(ctx, "env vars diff, trigger a restart of the app")
	} else {
		tflog.Debug(ctx, "no env vars diff, no application restart")
	}

	// A restart reuses the build cache and skips the build phase, env vars
	// consumed at build time need a rebuild to be taken into account
	triggerRebuild := false
	if rt, ok := any(plan).(RebuildTrigger); ok {
		triggerRebuild = envKeysDiffer(planEnvironment, stateEnvironment, rt.RebuildEnvKeys())
	}
	if triggerRebuild {
		tflog.Debug(ctx, "build time env vars diff, restart without build cache")
	}

	// Build UpdateReq
	updateReq := UpdateReq{
		ID:             stateRuntime.ID.ValueString(),
		Client:         resource.Client(),
		Organization:   resource.Organization(),
		Application:    wannabeApplication(runtime, instance),
		Environment:    planEnvironment,
		VHosts:         vhosts,
		Deployment:     config.ToDeployment(resource.GitAuth()),
		TriggerRestart: triggerRestart,
		TriggerRebuild: triggerRebuild,
	}

	// Call common Update function
	updatedApp, updateDiags := UpdateApp(ctx, updateReq)
	diags.Append(updateDiags...)

	// Sync response even if there were errors (app might be updated)
	if updatedApp != nil {
		runtime.FromApp(ctx, updatedApp, &diags)
		resolveUnknownCommit(runtime.Deployment, updatedApp.TargetCommit)
	} else {
		resolveUnknownCommit(runtime.Deployment, "")
	}

	// TCP redirection
	runtime.Redirection = SyncTCPRedirection(ctx, resource.Client(), resource.Organization(), stateRuntime.ID.ValueString(), runtime.Redirection, stateRuntime.Redirection, &diags)

	if !runtime.Tags.Equal(stateRuntime.Tags) {
		runtime.Tags = SyncTags(ctx, resource, stateRuntime.ID.ValueString(), runtime.Tags, stateRuntime.Tags, &diags)
	}

	return diags
}

// envKeysDiffer reports whether any of keys is added, removed or changed between both environments
func envKeysDiffer(planEnv, stateEnv map[string]string, keys []string) bool {
	for _, key := range keys {
		planValue, inPlan := planEnv[key]
		stateValue, inState := stateEnv[key]
		if inPlan != inState || planValue != stateValue {
			return true
		}
	}
	return false
}

// restartApp redeploys the application. Without withoutCache, the platform restores
// the build cache and skips the build phase.
func restartApp(ctx context.Context, cc *client.Client, organisationID, applicationID string, withoutCache bool) client.Response[models.DeploySuccessMessage] {
	var opts []applications.Option
	if withoutCache {
		opts = append(opts, applications.WithUsecache("no"))
	}
	return pkg.SDK(cc).V2().Organisations().ID(organisationID).Applications().Appid(applicationID).Instances().Redeployorganisationapplication(ctx, opts...)
}
