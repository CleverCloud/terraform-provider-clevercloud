package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sirupsen/logrus"
	"go.clever-cloud.dev/client"
)

func (p *Provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	p.configured.Do(func() {
		var config ProviderData

		tflog.Debug(ctx, "configure provider...")

		diags := req.Config.Get(ctx, &config)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
		if os.Getenv("DEBUG") == "true" {
			logger.SetLevel(logrus.DebugLevel)
		}

		p.Organisation = config.Organisation.Value
		p.cc = client.New(client.WithAutoOauthConfig(), client.WithLogger(logger))

		selfRes := client.Get[map[string]interface{}](ctx, p.cc, "/v2/self")
		if selfRes.HasError() {
			resp.Diagnostics.AddError("invalid CleverCloud Client configuration", selfRes.Error().Error())
			return
		}

		tflog.Debug(ctx, "provider configured")
	})
}
