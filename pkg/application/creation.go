package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	ID           string
	Client       *client.Client
	Organization string
	Application  tmp.UpdateAppReq
	Environment  map[string]string
	VHosts       []string
	Deployment   *Deployment
	Dependencies []string
}

type Deployment struct {
	Repository     string
	Commit         *string
	User, Password *string
	PrivateSSHKey  *string
}

type CreateRes struct {
	Application tmp.CreatAppResponse
}

func CreateApp(ctx context.Context, req CreateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Application
	res := &CreateRes{}

	appRes := tmp.CreateApp(ctx, req.Client, req.Organization, req.Application)
	if appRes.HasError() {
		diags.AddError("failed to create application", appRes.Error().Error())
		tflog.Error(ctx, "failed to create app", map[string]interface{}{"error": appRes.Error().Error(), "payload": fmt.Sprintf("%+v", req.Application)})
		return nil, diags
	}

	res.Application = *appRes.Payload()

	// Environment
	envRes := tmp.UpdateAppEnv(ctx, req.Client, req.Organization, res.Application.ID, req.Environment)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
	}

	// VHosts
	for _, vhost := range req.VHosts {
		addVhostRes := tmp.AddAppVHost(ctx, req.Client, req.Organization, res.Application.ID, vhost)
		if addVhostRes.HasError() {
			diags.AddError("failed to add additional vhost", addVhostRes.Error().Error())
		}
	}

	// Git Deployment
	if req.Deployment != nil {
		diags.Append(gitDeploy(ctx, *req.Deployment, req.Client, res.Application.DeployURL)...)
	}

	// Dependencies
	for _, dependency := range req.Dependencies {
		// TODO: support another apps as dependency

		depRes := tmp.AddAppLinkedAddons(ctx, req.Client, req.Organization, res.Application.ID, dependency)
		if depRes.HasError() {
			tflog.Error(ctx, "ERROR: "+dependency, map[string]interface{}{"err": depRes.Error().Error()})
			diags.AddError("failed to add dependency", depRes.Error().Error())
		}
	}

	return res, diags
}

func UpdateApp(ctx context.Context, req UpdateReq) (*CreateRes, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Application
	res := &CreateRes{}

	appRes := tmp.UpdateApp(ctx, req.Client, req.Organization, req.ID, req.Application)
	if appRes.HasError() {
		diags.AddError("failed to update application", appRes.Error().Error())
		tflog.Error(ctx, "failed to update app", map[string]interface{}{"error": appRes.Error().Error(), "payload": fmt.Sprintf("%+v", req.Application)})
		return nil, diags
	}

	res.Application = *appRes.Payload()

	// Environment
	envRes := tmp.UpdateAppEnv(ctx, req.Client, req.Organization, res.Application.ID, req.Environment)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
	}

	// VHosts
	for _, vhost := range req.VHosts {
		addVhostRes := tmp.AddAppVHost(ctx, req.Client, req.Organization, res.Application.ID, vhost)
		if addVhostRes.HasError() {
			diags.AddError("failed to add additional vhost", addVhostRes.Error().Error())
		}
	}
	// TODO: old vhost need to be cleaned

	// Git Deployment
	if req.Deployment != nil {
		diags.Append(gitDeploy(ctx, *req.Deployment, req.Client, res.Application.DeployURL)...)
	}

	// Dependencies
	for _, dependency := range req.Dependencies {
		// TODO: support another apps as dependency

		depRes := tmp.AddAppLinkedAddons(ctx, req.Client, req.Organization, res.Application.ID, dependency)
		if depRes.HasError() {
			tflog.Error(ctx, "ERROR: "+dependency, map[string]interface{}{"err": depRes.Error().Error()})
			diags.AddError("failed to add dependency", depRes.Error().Error())
		}
	}
	// TODO: unlink unneeded deps

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
