package pulsar

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_pulsar.
// See CONTRIBUTING.md § "API → state mapping".
//
// readRetention is deliberately not here: it talks to the Pulsar admin API, and
// a mapper makes no API call.

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (p *Pulsar) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) *Pulsar {
	if p == nil || addon == nil {
		return p
	}

	p.Name = pkg.FromStr(addon.Name)
	p.Region = pkg.FromStr(addon.Region)

	return p
}

// FromPulsar maps the product view: the tenant, namespace and token.
func (p *Pulsar) FromPulsar(ctx context.Context, api *tmp.Pulsar, diags *diag.Diagnostics) *Pulsar {
	if p == nil || api == nil {
		return p
	}

	p.Tenant = pkg.FromStr(api.Tenant)
	p.Namespace = pkg.FromStr(api.Namespace)
	p.Token = pkg.FromStr(api.Token)

	return p
}

// FromCluster assembles the two endpoint URLs. The scheme follows the presence
// of a TLS port: the cluster answers both, and only one of each pair is live.
func (p *Pulsar) FromCluster(ctx context.Context, cluster *tmp.PulsarCluster, diags *diag.Diagnostics) *Pulsar {
	if p == nil || cluster == nil {
		return p
	}

	if cluster.PulsarTLSPort != 0 {
		p.BinaryURL = pkg.FromStr(fmt.Sprintf("pulsar+ssl://%s:%d", cluster.URL, cluster.PulsarTLSPort))
	} else {
		p.BinaryURL = pkg.FromStr(fmt.Sprintf("pulsar://%s:%d", cluster.URL, cluster.PulsarPort))
	}

	if cluster.WebTLSPort != 0 {
		p.HTTPUrl = pkg.FromStr(fmt.Sprintf("https://%s:%d", cluster.URL, cluster.WebTLSPort))
	} else {
		p.HTTPUrl = pkg.FromStr(fmt.Sprintf("http://%s:%d", cluster.URL, cluster.WebPort))
	}

	return p
}
