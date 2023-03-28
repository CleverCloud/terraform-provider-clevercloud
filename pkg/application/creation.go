package application

import (
	"context"

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
}

type CreateRes struct {
	Application tmp.CreatAppResponse
}

func CreateApp(ctx context.Context, req CreateReq, diags diag.Diagnostics) *CreateRes {
	// Application
	res := &CreateRes{}

	appRes := tmp.CreateApp(ctx, req.Client, req.Organization, req.Application)
	if appRes.HasError() {
		diags.AddError("failed to create application", appRes.Error().Error())
		tflog.Error(ctx, "failed to create app", map[string]interface{}{"error": appRes.Error().Error()})
		return nil
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

	return res
}
