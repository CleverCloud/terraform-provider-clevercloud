package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_elasticsearch.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view. It is the only source of name, plan
// and region: tmp.Elasticsearch carries none of them.
func (es *Elasticsearch) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Elasticsearch {
	if es == nil || addon == nil {
		return es
	}

	es.Name = pkg.FromStr(addon.Name)
	es.Plan = pkg.FromStr(addon.Plan.Slug)
	es.Region = pkg.FromStr(addon.Region)

	return es
}

// FromElasticsearch maps the product view: the toggles, the credentials each one
// unlocks, and the plugin list.
//
// encryption, kibana and apm are Optional + Computed + Default(false), and
// kibana and apm additionally carry RequiresReplace: left null in state after an
// import their defaults create a diff, and the next plan destroys and recreates
// the add-on. They are therefore assigned on every read, absent from the API or
// not (#452).
func (es *Elasticsearch) FromElasticsearch(ctx context.Context, api *tmp.Elasticsearch, diags *diag.Diagnostics) *Elasticsearch {
	if es == nil || api == nil {
		return es
	}

	es.Host = pkg.FromStr(api.Host)
	es.User = pkg.FromStr(api.User)
	es.Password = pkg.FromStr(api.Password)

	// The schema holds the major version alone; the API answers a full semver.
	if api.Version != "" {
		if v, err := semver.NewVersion(api.Version); err == nil {
			es.Version = pkg.FromStr(fmt.Sprintf("%d", v.Major()))
		} else if parts := strings.Split(api.Version, "."); len(parts) > 0 {
			es.Version = pkg.FromStr(parts[0])
		} else {
			es.Version = pkg.FromStr(api.Version)
		}
	}

	features := pkg.FeaturesOf(api.Features, func(f tmp.ElasticsearchFeature) (string, bool) {
		return f.Name, f.Enabled
	})
	features.Or(&es.Encryption, "encryption", false)
	features.Or(&es.Kibana, "kibana", false)
	features.Or(&es.Apm, "apm", false)

	// The credentials only exist while their service is on, so they are reset
	// first and filled in only for the services the API reports as enabled.
	es.KibanaUser = basetypes.NewStringNull()
	es.KibanaPassword = basetypes.NewStringNull()
	es.KibanaHost = basetypes.NewStringNull()
	if features.Enabled("kibana") {
		es.KibanaUser = pkg.FromStr(api.KibanaUser)
		es.KibanaPassword = pkg.FromStr(api.KibanaPassword)
	}
	if api.KibanaHost != nil {
		es.KibanaHost = pkg.FromStr(*api.KibanaHost)
	}

	es.ApmUser = basetypes.NewStringNull()
	es.ApmPassword = basetypes.NewStringNull()
	es.ApmToken = basetypes.NewStringNull()
	es.ApmHost = basetypes.NewStringNull()
	if features.Enabled("apm") {
		es.ApmUser = pkg.FromStr(api.ApmUser)
		es.ApmPassword = pkg.FromStr(api.ApmPassword)
		es.ApmToken = pkg.FromStr(api.ApmAuthToken)
		// Guarded, unlike before: the API has been seen to report apm enabled
		// with no host yet while the service boots, and dereferencing it then
		// panicked the provider.
		if api.ApmHost != nil {
			es.ApmHost = pkg.FromStr(*api.ApmHost)
		}
	}

	es.Plugins = basetypes.NewSetNull(types.StringType)
	if len(api.Plugins) > 0 {
		es.Plugins = pkg.FromSetString(api.Plugins, diags)
	}

	return es
}
