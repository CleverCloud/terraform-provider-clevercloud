package actions

import (
	"context"
	_ "embed"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func RebootApplication() action.Action {
	return &ActionRebootApplication{}
}

type ActionRebootApplication struct {
	provider.Provider
}

type rebootApplication struct {
	ApplicationID types.String `tfsdk:"application_id"`
}

func (ar *ActionRebootApplication) Configure(ctx context.Context, req action.ConfigureRequest, res *action.ConfigureResponse) {
	tflog.Debug(ctx, "Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if provider, ok := req.ProviderData.(provider.Provider); ok {
		ar.Provider = provider
	}

	tflog.Debug(ctx, "Configured", map[string]any{"org": ar.Organization()})
}

//go:embed doc.md
var actionRebootApplicationDoc string

func (ar *ActionRebootApplication) Schema(ctx context.Context, req action.SchemaRequest, res *action.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: actionRebootApplicationDoc,
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "Application to reboot",
				Validators: []validator.String{
					pkg.NewStringValidator(
						"must be an application ID",
						func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() {
								res.Diagnostics.AddError("cannot be null", "application id is null")
							} else if req.ConfigValue.IsUnknown() {
								return
							}

							if !strings.HasPrefix(req.ConfigValue.ValueString(), "app_") {
								res.Diagnostics.AddError("expect a valid application ID", "ID don't start with 'app_'")
							}
						},
					),
				},
			},
		},
	}
}

func (ar *ActionRebootApplication) Metadata(ctx context.Context, req action.MetadataRequest, res *action.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_application_reboot"
}

func (ar *ActionRebootApplication) Invoke(ctx context.Context, req action.InvokeRequest, res *action.InvokeResponse) {
	tflog.Debug(ctx, "Invoke application_reboot", map[string]any{
		"config": req.Config,
	})

	cfg := helper.From[rebootApplication](ctx, req.Config, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	rebootRes := tmp.RebootApp(ctx, ar.Client(), ar.Organization(), cfg.ApplicationID.ValueString())
	if rebootRes.HasError() {
		res.Diagnostics.AddError("failed to reboot application", rebootRes.Error().Error())
	}
	reboot := rebootRes.Payload()

	res.SendProgress(action.InvokeProgressEvent{Message: "Restarting application"})

	stateC := WatchDeployment(
		ctx,
		ar.Client(),
		ar.Organization(),
		cfg.ApplicationID.ValueString(),
		reboot.DeploymentID,
		&res.Diagnostics,
	)

	for deployment := range stateC {
		switch deployment.State {
		case "WIP":
			res.SendProgress(action.InvokeProgressEvent{Message: "New deployment has started..."})
		case "OK":
			res.SendProgress(action.InvokeProgressEvent{Message: "Successfully rebooted"})
		case "FAIL":
			res.Diagnostics.AddError("failed to reboot application", "see application logs for details")
		}
	}
}

/**
 * rien => pas encore start
 * WIP => en train de reboot
 * OK => finis
 * FAIL => error
 */
func WatchDeployment(
	ctx context.Context,
	client *client.Client,
	organisation,
	application,
	deployment string,
	diags *diag.Diagnostics,
) <-chan tmp.DeploymentResponse {
	out := make(chan tmp.DeploymentResponse, 1)
	var lastState tmp.DeploymentResponse

	go func() {
		for {
			if ctx.Err() != nil { // context Done
				return
			}

			deployRes := tmp.GetDeployment(ctx, client, organisation, application, deployment)
			if deployRes.HasError() {
				tflog.Error(ctx, "failed to get deployment status", map[string]any{
					"error": deployRes.Error().Error(),
				})
				time.Sleep(250 * time.Millisecond)
				continue
			}

			deploy := deployRes.Payload()
			tflog.Info(ctx, "deploy status", map[string]any{
				"state":  deploy.State,
				"date":   deploy.Date,
				"action": deploy.Action,
			})

			if !lastState.Equal(deploy) {
				out <- *deploy
				lastState = *deploy
			}

			// Final states
			if deploy.State == "FAIL" || deploy.State == "OK" {
				close(out)
				return
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return out
}
