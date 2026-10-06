package keycloak

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

// mapping.go holds every API-to-state mapper for clevercloud_keycloak.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (kc *Keycloak) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Keycloak {
	if kc == nil || addon == nil {
		return kc
	}

	kc.Name = pkg.FromStr(addon.Name)
	kc.Region = pkg.FromStr(addon.Region)

	return kc
}

// FromKeycloak maps the product view: the access URL, the initial credentials,
// the version and the bucket backing it.
//
// access_domain lives in the payload's own env map rather than in a field of its
// own, which is why this resource has no FromEnv: there is no add-on env call to
// make, the variable arrives inside the product view.
func (kc *Keycloak) FromKeycloak(ctx context.Context, api *models.Keycloak, diags *diag.Diagnostics) *Keycloak {
	if kc == nil || api == nil {
		return kc
	}

	kc.Host = pkg.FromStr(api.AccessURL)
	kc.AdminUsername = pkg.FromStr(api.InitialCredentials.User)
	kc.AdminPassword = pkg.FromStr(api.InitialCredentials.Password)
	kc.Version = pkg.FromStr(api.Version)
	kc.AccessDomain = pkg.FromStr(api.EnvVars["CC_KEYCLOAK_HOSTNAME"])
	kc.FSBucketID = types.StringPointerValue(api.Resources.FsbucketID)

	return kc
}
