package addon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.dev/client"
)

// AddonResource interface defines methods required by resources to use generic addon CRUD operations
type AddonResource interface {
	provider.Provider
	GetSlug() string
}

// AddonPlan interface defines methods required by plan types to use generic addon CRUD operations
type AddonPlan interface {
	// GetCommonPtr returns a pointer to the CommonAttributes struct for modification
	GetCommonPtr() *CommonAttributes

	// GetAddonOptions returns provider-specific options for addon creation (e.g. encryption, version)
	GetAddonOptions() map[string]string

	// SetFromResponse reads provider-specific fields (host, port, password...) from the API.
	// cc is the API client, org is the organisation ID, addonID is the addon_xxx ID.
	SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics)

	// SetDefaults sets default values for optional boolean/computed fields after creation
	SetDefaults()
}
