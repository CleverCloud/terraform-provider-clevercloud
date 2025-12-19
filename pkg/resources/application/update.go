package application

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// UpdateReq represents the request structure for updating an application
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

// UpdateApp handles the low-level API calls for updating an application
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
		return res, diags
	}

	// VHosts
	SyncVHostsOnUpdate(ctx, req.Client, req.Organization, req.VHosts, &diags, res.Application.ID)

	// This is dirty, but we need a refresh
	vhostsRes := tmp.GetAppVhosts(ctx, req.Client, req.Organization, res.Application.ID)
	if vhostsRes.HasError() {
		diags.AddError("failed to get application vhosts", vhostsRes.Error().Error())
		return res, diags
	}
	res.Application.Vhosts = *vhostsRes.Payload()

	// Dependencies - sync using Set.Difference pattern (add new, remove old)
	// TODO: support apps as dependencies
	// see: https://www.clever.cloud/developers/doc/administrate/service-dependencies/
	dependenciesWithAddonIDs, err := tmp.RealIDsToAddonIDs(ctx, req.Client, req.Organization, req.Dependencies...)
	if err != nil {
		diags.AddError("failed to get dependencies add-on IDs", err.Error())
		return res, diags
	}

	SyncDependencies(ctx, req.Client, req.Organization, res.Application.ID, dependenciesWithAddonIDs, &diags)
	if diags.HasError() {
		return res, diags
	}

	// Git Deployment (when commit change)
	gitDeployed := false
	if req.Deployment != nil {
		GitDeploy(ctx, req.Deployment, res.Application.DeployURL, &diags)
		if diags.HasError() {
			return res, diags
		}
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
				return res, diags
			}
		}
	}

	return res, diags
}

// Update centralizes the common Update logic for all application runtimes
func Update[T RuntimePlan](ctx context.Context, resource RuntimeResource, plan, state T) diag.Diagnostics {
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

	// Extract vhosts and dependencies from plan
	vhosts := plan.VHostsAsStrings(ctx, &diags)
	dependencies := plan.DependenciesAsString(ctx, &diags)
	if diags.HasError() {
		return diags
	}

	// Get runtime pointer to access common fields
	runtime := plan.GetRuntimePtr()
	stateRuntime := state.GetRuntimePtr()

	// Build UpdateReq
	updateReq := UpdateReq{
		ID:           stateRuntime.ID.ValueString(),
		Client:       resource.Client(),
		Organization: resource.Organization(),
		Application: tmp.UpdateAppReq{
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
		Environment:    planEnvironment,
		VHosts:         vhosts,
		Dependencies:   dependencies,
		Deployment:     plan.ToDeployment(resource.GitAuth()),
		TriggerRestart: !reflect.DeepEqual(planEnvironment, stateEnvironment),
	}

	// Call common Update function
	updatedApp, updateDiags := UpdateApp(ctx, updateReq)
	diags.Append(updateDiags...)

	// Update VHosts even if there were errors (app might be updated)
	if updatedApp != nil {
		runtime.VHosts = helper.VHostsFromAPIHosts(ctx, updatedApp.Application.Vhosts.AsString(), runtime.VHosts, &diags)
	}

	return diags
}
