package pulsar

import (
	"context"
	_ "embed"

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

	BinaryURL            types.String `tfsdk:"binary_url"`
	HTTPUrl              types.String `tfsdk:"http_url"`
	Tenant               types.String `tfsdk:"tenant"`
	Namespace            types.String `tfsdk:"namespace"`
	Token                types.String `tfsdk:"token"`
	RetentionSize        types.Int64  `tfsdk:"retention_size"`
	RetentionTime        types.Int64  `tfsdk:"retention_time"`
	OffloadThresholdSize types.Int64  `tfsdk:"offload_threshold_size"`
}

//go:embed doc.md
var resourcePulsarDoc string

func (r ResourcePulsar) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourcePulsarDoc,
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":                   schema.StringAttribute{Required: true, MarkdownDescription: "Name of the Pulsar"},
			"region":                 schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Geographical region where the data will be stored", Default: stringdefault.StaticString("par")},
			"binary_url":             schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar native protocol address"},
			"http_url":               schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar REST API address"},
			"tenant":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar tenant"},
			"namespace":              schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar namespace"},
			"token":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Pulsar authentication token"},
			"retention_size":         schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar retention size in megabytes"},
			"retention_time":         schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar retention time in days"},
			"offload_threshold_size": schema.Int64Attribute{Optional: true, MarkdownDescription: "Pulsar offload size in megabytes"},
		},
	}
}

// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade#implementing-state-upgrade-support
func (r ResourcePulsar) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
