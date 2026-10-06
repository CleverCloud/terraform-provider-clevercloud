package networkgroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/sdk/models"
)

// mapping.go holds every API-to-state mapper for clevercloud_networkgroup.
// See CONTRIBUTING.md § "API → state mapping".

// FromNetworkGroup maps the network-group view.
//
// For the optional fields (description, tags), the rule is:
//   - state null + API empty → keep null (user never set it, nothing to sync)
//   - otherwise              → sync from API (state was set, or API returned a value)
func (ng *Networkgroup) FromNetworkGroup(ctx context.Context, api *models.NetworkGroup1, diags *diag.Diagnostics) *Networkgroup {
	if ng == nil || api == nil {
		return ng
	}

	ng.Name = pkg.FromStrMaxLen(api.Label)

	apiDescEmpty := api.Description == nil || *api.Description == ""
	if !ng.Description.IsNull() || !apiDescEmpty {
		ng.Description = basetypes.NewStringPointerValue(api.Description)
	}

	apiTagsEmpty := len(api.Tags) == 0
	if !ng.Tags.IsNull() || !apiTagsEmpty {
		ng.Tags = pkg.FromSetString(api.Tags, diags)
	}

	ng.Network = pkg.FromStr(api.NetworkIP)

	return ng
}
