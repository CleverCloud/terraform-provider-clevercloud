package v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func CreateApplication(ctx context.Context, client *client.Client, orga string, app tmp.CreateAppRequest) (*tmp.CreatAppResponse, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if app.BuildFlavor != "" {
		app.SeparateBuild = true
	}

	appRes := tmp.CreateApp(ctx, client, orga, app)
	if appRes.HasError() {
		diags.AddError("failed to create application", appRes.Error().Error())
	}
	return appRes.Payload(), diags
}

func UpsertEnvironment(ctx context.Context, client *client.Client, orga string, appID string, env map[string]string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	// Environment
	envRes := tmp.UpdateAppEnv(ctx, client, orga, appID, env)
	if envRes.HasError() {
		diags.AddError("failed to configure application environment", envRes.Error().Error())
	}
	return diags
}

func UpsertDependencies(ctx context.Context, client *client.Client, orga string, appID string, deps []string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	dependenciesWithAddonIDs, err := tmp.RealIDsToAddonIDs(ctx, client, orga, deps...)
	if err != nil {
		diags.AddError("failed to get dependencies addon IDs", err.Error())
	}

	for _, dependency := range dependenciesWithAddonIDs {
		depRes := tmp.AddAppLinkedAddons(ctx, client, orga, appID, dependency)
		if depRes.HasError() {
			diags.AddError("failed to add dependency", depRes.Error().Error())
		}
	}

	// TODO: unlink unneeded deps
	return diags
}

func UpsertVHosts(ctx context.Context, client *client.Client, orga string, appID string, vhosts []string) diag.Diagnostics {
	diags := diag.Diagnostics{}
	application.UpdateVhosts(ctx, client, orga, vhosts, &diags, appID)
	return diags
}

func Deploy(ctx context.Context, client *client.Client, orga string, appID string, deployment *application.Deployment) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if deployment == nil {
		return diags
	}

	diags.Append(application.GitDeploy(ctx, *deployment, client, appID)...)
	return diags
}

func GetBuildFlavor(app *tmp.CreatAppResponse) types.String {
	if app == nil || !app.SeparateBuild {
		return types.StringNull()
	}

	return types.StringValue(app.BuildFlavor.Name)
}
