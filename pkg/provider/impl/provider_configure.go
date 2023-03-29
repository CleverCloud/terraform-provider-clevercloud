package impl

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ProviderData

	tflog.Info(ctx, "configure provider...")

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.organization = config.Organisation.ValueString()

	// Allow to get creds from CLI config directory or by injected variables
	if config.Secret.IsUnknown() ||
		config.Token.IsUnknown() ||
		config.Secret.IsNull() ||
		config.Token.IsNull() {
		p.cc = client.New(client.WithAutoOauthConfig())
	} else {
		p.cc = client.New(client.WithUserOauthConfig(
			config.Token.ValueString(),
			config.Secret.ValueString(),
		))
	}

	selfRes := client.Get[map[string]interface{}](ctx, p.cc, "/v2/self")
	if selfRes.HasError() {
		if selfRes.StatusCode() == 401 || selfRes.StatusCode() == 403 {
			resp.Diagnostics.AddError("invalid CleverCloud Client configuration", selfRes.Error().Error())
		} else {
			resp.Diagnostics.AddError(
				"Clever Cloud is not available :/",
				fmt.Sprintf(
					"you can contact the Clever Cloud support with the next Request ID: '%s'",
					selfRes.SozuID(),
				))
		}
		return
	}

	// We pass the full provider to the children resources
	resp.DataSourceData = p
	resp.ResourceData = p

	tflog.Info(ctx, "provider configured")
}
