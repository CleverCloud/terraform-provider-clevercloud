package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
)

type DataSourceConfigurer struct {
	provider.Provider
}

func (c *DataSourceConfigurer) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Debug(ctx, "Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if provider, ok := req.ProviderData.(provider.Provider); ok {
		c.Provider = provider
	}

	tflog.Debug(ctx, "Configured", map[string]any{"org": c.Organization()})
}
