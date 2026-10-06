package mongodb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_mongodb.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view. It is the only source of name, plan,
// region and creation_date: tmp.MongoDB carries none of them.
func (mg *MongoDB) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *MongoDB {
	if mg == nil || addon == nil {
		return mg
	}

	mg.Name = pkg.FromStr(addon.Name)
	mg.Plan = pkg.FromStr(addon.Plan.Slug)
	mg.Region = pkg.FromStr(addon.Region)
	mg.CreationDate = pkg.FromI(addon.CreationDate)

	return mg
}

// FromMongoDB maps the product view: connection details and the feature toggles.
//
// The feature list is sparse and both toggles are Computed, so an absent one
// resolves to its fallback and never to a null — a null is what left an
// imported add-on with a non-empty plan (#404).
func (mg *MongoDB) FromMongoDB(ctx context.Context, api *tmp.MongoDB, diags *diag.Diagnostics) *MongoDB {
	if mg == nil || api == nil {
		return mg
	}

	mg.Host = pkg.FromStr(api.Host)
	mg.Port = pkg.FromI(api.Port)
	mg.User = pkg.FromStr(api.User)
	mg.Password = pkg.FromStr(api.Password)
	mg.Database = pkg.FromStr(api.Database)
	mg.Uri = pkg.FromStr(api.Uri())

	features := pkg.FeaturesOf(api.Features, func(f tmp.MongoDBFeature) (string, bool) {
		return f.Name, f.Enabled
	})
	features.Or(&mg.Encryption, "encryption", false)
	features.Or(&mg.DirectHostOnly, "direct-host-only", false)

	return mg
}
