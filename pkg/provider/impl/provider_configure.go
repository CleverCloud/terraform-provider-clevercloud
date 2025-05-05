package impl

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
)

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ProviderData

	tflog.Debug(ctx, "configure provider...")

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sentry, on the higher CONFIGURABLE level
	if config.ErrorReports.ValueBool() { // implicit == true
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              "https://9f1c17cd85db40f5a1991aefcc182944@glitchtip.corp.clever-cloud.com/35",
			AttachStacktrace: true,
			Release:          pkg.Version,
			Dist:             req.TerraformVersion,
			Tags: map[string]string{
				"endpoint": config.Endpoint.ValueString(),
				"owner":    config.Organisation.ValueString(),
			},
		})
		if err != nil {
			tflog.Warn(ctx, "failed to setup sentry", map[string]interface{}{"error": err.Error()})
		}
		defer sentry.Flush(2 * time.Second)
	}

	if config.Organisation.IsUnknown() || config.Organisation.IsNull() {
		p.organization = os.Getenv("CC_ORGANISATION")
	} else {
		p.organization = config.Organisation.ValueString()
	}
	if p.organization == "" {
		resp.Diagnostics.AddError("Invalid provider configuration", "Organisation should be set by either the organisation parameter or by the CC_ORGANISATION environment variable")
		return
	}

	// Allow to get creds from CLI config directory or by injected variables
	var clientOptions []func(*client.Client)
	if !config.Endpoint.IsUnknown() && !config.Endpoint.IsNull() && config.Endpoint.ValueString() != "" {
		clientOptions = append(clientOptions, client.WithEndpoint(config.Endpoint.ValueString()))
	}

	// New branch: allow setting all OAuth1 params
	if !config.ConsumerKey.IsUnknown() && !config.ConsumerKey.IsNull() && config.ConsumerKey.ValueString() != "" &&
		!config.ConsumerSecret.IsUnknown() && !config.ConsumerSecret.IsNull() && config.ConsumerSecret.ValueString() != "" &&
		!config.Token.IsUnknown() && !config.Token.IsNull() && config.Token.ValueString() != "" &&
		!config.Secret.IsUnknown() && !config.Secret.IsNull() && config.Secret.ValueString() != "" {
		clientOptions = append(clientOptions, client.WithOauthConfig(
			config.ConsumerKey.ValueString(),
			config.ConsumerSecret.ValueString(),
			config.Token.ValueString(),
			config.Secret.ValueString(),
		))
		p.cc = client.New(clientOptions...)
	} else if config.Secret.IsUnknown() ||
		config.Token.IsUnknown() ||
		config.Secret.IsNull() ||
		config.Token.IsNull() {
		clientOptions = append(clientOptions, client.WithAutoOauthConfig())
		p.cc = client.New(clientOptions...)
	} else {
		clientOptions = append(clientOptions, client.WithUserOauthConfig(
			config.Token.ValueString(),
			config.Secret.ValueString(),
		))
		p.cc = client.New(clientOptions...)
	}

	selfRes := client.Get[map[string]any](ctx, p.cc, "/v2/self")
	if selfRes.HasError() {
		endpoint := config.Endpoint.ValueString()
		tflog.Debug(ctx, fmt.Sprintf("CleverCloud client endpoint=%q", endpoint))
		if selfRes.StatusCode() == 401 || selfRes.StatusCode() == 403 {
			resp.Diagnostics.AddError("invalid CleverCloud Client configuration", selfRes.Error().Error())
		} else {
			resp.Diagnostics.AddError(
				"Unknown error from Clever Cloud",
				fmt.Sprintf(
					"Status %d, contact the Clever Cloud support with the next Request ID: '%s'\nError: %s",
					selfRes.StatusCode(), selfRes.SozuID(), selfRes.Error().Error(),
				))
		}
		return
	}

	// We pass the full provider to the children resources
	resp.DataSourceData = p
	resp.ResourceData = p

	tflog.Debug(ctx, "provider configured")
}
