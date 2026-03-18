package pulsar

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsaradmin"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type Pulsar struct {
	addon.CommonAttributes

	BinaryURL types.String `tfsdk:"binary_url"`
	HTTPUrl   types.String `tfsdk:"http_url"`
	Tenant    types.String `tfsdk:"tenant"`
	Namespace types.String `tfsdk:"namespace"`
	Token     types.String `tfsdk:"token"`

	RetentionSize   types.Int64 `tfsdk:"retention_size"`
	RetentionPeriod types.Int64 `tfsdk:"retention_period"`
}

func (p *Pulsar) GetCommonPtr() *addon.CommonAttributes {
	return &p.CommonAttributes
}

func (p *Pulsar) GetAddonOptions() map[string]string {
	return map[string]string{}
}

func (p *Pulsar) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	pulsarRes := tmp.GetPulsar(ctx, cc, org, p.ID.ValueString())
	if pulsarRes.HasError() {
		diags.AddError("failed to get Pulsar", pulsarRes.Error().Error())
		return
	}
	pulsar := pulsarRes.Payload()
	readAddon(p, pulsar, diags)

	pulsarClusterRes := tmp.GetPulsarCluster(ctx, cc, pulsar.ClusterID)
	if pulsarClusterRes.HasError() {
		diags.AddError("failed to get Pulsar cluster", pulsarClusterRes.Error().Error())
		return
	}
	pulsarCluster := pulsarClusterRes.Payload()
	readCluster(p, pulsarCluster, diags)

	readRetention(ctx, p, diags)
}

func (p *Pulsar) SetDefaults() {
	// Pulsar has no optional fields requiring defaults
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
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// Single-plan addon: plan is computed, not user-specified
			"plan":             schema.StringAttribute{Computed: true},
			"binary_url":       schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar native protocol address"},
			"http_url":         schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar REST API address"},
			"tenant":           schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar tenant"},
			"namespace":        schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar namespace"},
			"token":            schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar authentication token", Sensitive: true},
			"retention_size":   schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar namespace retention policy in bytes"},
			"retention_period": schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar namespace retention policy in minutes"},
		}),
	}
}
