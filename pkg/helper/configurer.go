package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.dev/sdk"
)

type Configurer struct {
	provider.Provider
	SDK sdk.SDK
}

func (c *Configurer) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	if provider, ok := req.ProviderData.(provider.Provider); ok {
		c.Provider = provider
		c.SDK = sdk.NewSDK(sdk.WithClient(provider.Client()))
	}

	tflog.Debug(ctx, "Configured", map[string]any{"org": c.Organization()})
}

// Import resource
func (c *Configurer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}

// UpgradeState is the default for resources still at their first schema version.
// Bumping a schema Version requires the resource to override this method with an
// upgrader for every prior version, otherwise Terraform refuses to load any state
// holding an older instance. TestResources_stateUpgradersCoverEveryPriorVersion
// in pkg/registry enforces it.
//
// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (c *Configurer) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
