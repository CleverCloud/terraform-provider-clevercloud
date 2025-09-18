package pulsar

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsaradmin"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourcePulsar) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
