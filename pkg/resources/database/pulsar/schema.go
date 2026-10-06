package pulsar

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsaradmin"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type Pulsar struct {
	ID types.String `tfsdk:"id"`

	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`

	BinaryURL types.String `tfsdk:"binary_url"`
	HTTPUrl   types.String `tfsdk:"http_url"`
	Tenant    types.String `tfsdk:"tenant"`
	Namespace types.String `tfsdk:"namespace"`
	Token     types.String `tfsdk:"token"`

	RetentionSize   types.Int64 `tfsdk:"retention_size"`
	RetentionPeriod types.Int64 `tfsdk:"retention_period"`
}

func (p *Pulsar) TenantAndNamespace() string {
	return fmt.Sprintf("%s/%s", p.Tenant.ValueString(), p.Namespace.ValueString())
}

func (p *Pulsar) AdminClient() (pulsaradmin.Client, error) {
	cfg := &pulsaradmin.Config{WebServiceURL: p.HTTPUrl.ValueString(), Token: p.Token.ValueString()}
	return pulsaradmin.NewClient(cfg)
}

//go:embed doc.md
var resourcePulsarDoc string

func (r ResourcePulsar) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePulsarDoc,
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":             schema.StringAttribute{Required: true, MarkdownDescription: "Name of the Pulsar"},
			"region":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Geographical region where the data will be stored", Default: stringdefault.StaticString("par")},
			"binary_url":       schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar native protocol address"},
			"http_url":         schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar REST API address"},
			"tenant":           schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar tenant"},
			"namespace":        schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar namespace"},
			"token":            schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar authentication token", Sensitive: true},
			"retention_size":   schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar namespace retention policy in bytes"},
			"retention_period": schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar namespace retention policy in minutes"},
		},
	}
}

// The API-to-state mappers for clevercloud_pulsar follow.
// See CONTRIBUTING.md § "API → state mapping".
//
// readRetention is deliberately not here: it talks to the Pulsar admin API, and
// a mapper makes no API call.

// FromAddon maps the generic add-on view, the only source of name and region
// for this resource.
func (p *Pulsar) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if p == nil || addon == nil {
		return
	}

	p.Name = pkg.FromStr(addon.Name)
	p.Region = pkg.FromStr(addon.Region)
}

// FromPulsar maps the product view: the tenant, namespace and token.
func (p *Pulsar) FromPulsar(ctx context.Context, api *tmp.Pulsar, diags *diag.Diagnostics) {
	if p == nil || api == nil {
		return
	}

	p.Tenant = pkg.FromStr(api.Tenant)
	p.Namespace = pkg.FromStr(api.Namespace)
	p.Token = pkg.FromStr(api.Token)
}

// FromCluster assembles the two endpoint URLs. The scheme follows the presence
// of a TLS port: the cluster answers both, and only one of each pair is live.
func (p *Pulsar) FromCluster(ctx context.Context, cluster *tmp.PulsarCluster, diags *diag.Diagnostics) {
	if p == nil || cluster == nil {
		return
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
}
