package otoroshi

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_otoroshi.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view, the only source of name, region and
// creation_date for this resource.
func (o *Otoroshi) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Otoroshi {
	if o == nil || addon == nil {
		return o
	}

	o.Name = pkg.FromStr(addon.Name)
	o.Region = pkg.FromStr(addon.Region)
	o.CreationDate = pkg.FromI(addon.CreationDate)

	return o
}

// FromOtoroshi maps the product view: the access URL, the admin credentials and
// the API client, which only exists once the API is reachable.
//
// The entrypoint application the networkgroup sync needs is deliberately not
// mapped: it is not a state attribute, and the CRUD reads it off the payload
// itself — a mapper has no side effect.
func (o *Otoroshi) FromOtoroshi(ctx context.Context, api *tmp.OtoroshiInfo, diags *diag.Diagnostics) *Otoroshi {
	if o == nil || api == nil {
		return o
	}

	if api.API != nil {
		o.APIURL = pkg.FromStr(api.API.URL)
		o.APIClientID = pkg.FromStr(api.API.User)
		o.APIClientSecret = pkg.FromStr(api.API.Secret)
	}

	o.InitialAdminLogin = pkg.FromStr(api.Initialredentials.User)
	o.InitialAdminPassword = pkg.FromStr(api.Initialredentials.Passsword)
	o.URL = pkg.FromStr(api.AccessURL)

	return o
}
