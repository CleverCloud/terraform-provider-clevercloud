package drain

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

//go:embed doc.md
var resourceDrainDoc string

type Drain struct {
	ID         types.String `tfsdk:"id"`
	Kind       types.String `tfsdk:"kind"`
	ResourceID types.String `tfsdk:"resource_id"`
}

func (r ResourceDrain[T]) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceDrainDoc,
		Attributes: pkg.Merge(r.t.Attributes(), map[string]schema.Attribute{
			"kind": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "either LOG, ACCESSLOG or AUDITLOG",
				Default:             stringdefault.StaticString(string(tmp.DRAIN_KIND_LOG)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					pkg.NewStringEnumValidator(
						"Log kind",
						string(tmp.DRAIN_KIND_ACCESSLOG),
						string(tmp.DRAIN_KIND_AUDITLOG),
						string(tmp.DRAIN_KIND_LOG),
					),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Drain ID",
			},
			"resource_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Application or product ID which support logs drains",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		}),
	}
}

// DrainAttributes interface for common drain attributes
type DrainAttributes interface {
	Attributes() map[string]schema.Attribute
	ToSDKRecipient() models.DrainRecipient
	FromSDKRecipient(recipient models.DrainRecipientView)
	GetDrain() Drain
	SetDrain(common Drain)
}

// Datadog drain
type DatadogDrain struct {
	Drain
	URL types.String `tfsdk:"url"`
}

func (r DatadogDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Datadog log intake URL",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r DatadogDrain) ToSDKRecipient() models.DrainRecipient {
	return models.DatadogRecipient{
		Type: models.DatadogRecipientType,
		URL:  r.URL.ValueString(),
	}
}

func (r *DatadogDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.DatadogRecipientView); ok {
		r.URL = types.StringValue(v.URL)
	}
}

func (r DatadogDrain) GetDrain() Drain { return r.Drain }

func (r *DatadogDrain) SetDrain(common Drain) { r.Drain = common }

// New Relic drain
type NewRelicDrain struct {
	Drain
	URL    types.String `tfsdk:"url"`
	APIKey types.String `tfsdk:"api_key"`
}

func (r NewRelicDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "New Relic log API URL (e.g. https://log-api.newrelic.com/log/v1 or https://log-api.eu.newrelic.com/log/v1)",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"api_key": schema.StringAttribute{
			Required:            true,
			Sensitive:           true,
			MarkdownDescription: "New Relic API key used for ingestion",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r NewRelicDrain) ToSDKRecipient() models.DrainRecipient {
	return models.NewRelicRecipient{
		Type:   models.NewRelicRecipientType,
		URL:    r.URL.ValueString(),
		APIKey: r.APIKey.ValueString(),
	}
}

func (r *NewRelicDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.NewRelicRecipientView); ok {
		r.URL = types.StringValue(v.URL)
		// For sensitive attributes, preserve current state value since API doesn't return it
		// API key is not returned in the view, so we keep the existing value
	}
}

func (r NewRelicDrain) GetDrain() Drain { return r.Drain }

func (r *NewRelicDrain) SetDrain(common Drain) { r.Drain = common }

// Elasticsearch drain
type ElasticsearchDrain struct {
	Drain
	URL             types.String `tfsdk:"url"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	Index           types.String `tfsdk:"index"`
	TLSVerification types.String `tfsdk:"tls_verification"`
}

func (r ElasticsearchDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Elasticsearch URL",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"username": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Elasticsearch username",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"password": schema.StringAttribute{
			Required:            true,
			Sensitive:           true,
			MarkdownDescription: "Elasticsearch password",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"index": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Elasticsearch index name",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"tls_verification": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "TLS verification mode: DEFAULT or TRUSTFUL",
			Default:             stringdefault.StaticString("DEFAULT"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r ElasticsearchDrain) ToSDKRecipient() models.DrainRecipient {
	username := r.Username.ValueString()
	password := r.Password.ValueString()
	tlsMode := models.TLSMode(r.TLSVerification.ValueString())
	return models.ElasticsearchRecipient{
		Type:            models.ElasticsearchRecipientType,
		URL:             r.URL.ValueString(),
		Username:        &username,
		Password:        &password,
		Index:           r.Index.ValueString(),
		TLSVerification: &tlsMode,
	}
}

func (r *ElasticsearchDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.ElasticsearchRecipientView); ok {
		r.URL = types.StringValue(v.URL)
		if v.Username != nil {
			r.Username = types.StringValue(*v.Username)
		}
		// Password is sensitive and not returned in the view, so we keep the existing value
		r.Index = types.StringValue(v.Index)
		if v.TLSVerification != nil {
			r.TLSVerification = types.StringValue(string(*v.TLSVerification))
		}
	}
}

func (r ElasticsearchDrain) GetDrain() Drain { return r.Drain }

func (r *ElasticsearchDrain) SetDrain(drain Drain) { r.Drain = drain }

// Syslog UDP drain
type SyslogUDPDrain struct {
	Drain
	URL                         types.String `tfsdk:"url"`
	RFC5424StructuredDataParams types.String `tfsdk:"rfc5424_structured_data_parameters"`
}

func (r SyslogUDPDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Syslog UDP destination URL",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"rfc5424_structured_data_parameters": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "RFC5424 structured data parameters",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r SyslogUDPDrain) ToSDKRecipient() models.DrainRecipient {
	var params *string
	if !r.RFC5424StructuredDataParams.IsNull() && !r.RFC5424StructuredDataParams.IsUnknown() {
		p := r.RFC5424StructuredDataParams.ValueString()
		params = &p
	}
	return models.SyslogUDPRecipient{
		Type:                            models.SyslogUDPRecipientType,
		URL:                             r.URL.ValueString(),
		Rfc5424StructuredDataParameters: params,
	}
}

func (r *SyslogUDPDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.SyslogUDPRecipientView); ok {
		r.URL = types.StringValue(v.URL)
		if v.Rfc5424StructuredDataParameters != nil {
			r.RFC5424StructuredDataParams = types.StringValue(*v.Rfc5424StructuredDataParameters)
		}
		// Don't modify RFC5424 params if not returned (preserve current state)
	}
}

func (r SyslogUDPDrain) GetDrain() Drain { return r.Drain }

func (r *SyslogUDPDrain) SetDrain(common Drain) { r.Drain = common }

// Syslog TCP drain
type SyslogTCPDrain struct {
	Drain
	URL                         types.String `tfsdk:"url"`
	RFC5424StructuredDataParams types.String `tfsdk:"rfc5424_structured_data_parameters"`
}

func (r SyslogTCPDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Syslog TCP destination URL",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"rfc5424_structured_data_parameters": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "RFC5424 structured data parameters",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r SyslogTCPDrain) ToSDKRecipient() models.DrainRecipient {
	var params *string
	if !r.RFC5424StructuredDataParams.IsNull() && !r.RFC5424StructuredDataParams.IsUnknown() {
		p := r.RFC5424StructuredDataParams.ValueString()
		params = &p
	}
	return models.SyslogTCPRecipient{
		Type:                            models.SyslogTCPRecipientType,
		URL:                             r.URL.ValueString(),
		Rfc5424StructuredDataParameters: params,
	}
}

func (r *SyslogTCPDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.SyslogTCPRecipientView); ok {
		r.URL = types.StringValue(v.URL)
		if v.Rfc5424StructuredDataParameters != nil {
			r.RFC5424StructuredDataParams = types.StringValue(*v.Rfc5424StructuredDataParameters)
		}
		// Don't modify RFC5424 params if not returned (preserve current state)
	}
}

func (r SyslogTCPDrain) GetDrain() Drain { return r.Drain }

func (r *SyslogTCPDrain) SetDrain(common Drain) { r.Drain = common }

// HTTP drain
type HTTPDrain struct {
	Drain
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r HTTPDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "HTTP destination URL",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"username": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "HTTP basic auth username",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"password": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "HTTP basic auth password",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r HTTPDrain) ToSDKRecipient() models.DrainRecipient {
	var username, password *string
	if !r.Username.IsNull() && !r.Username.IsUnknown() {
		u := r.Username.ValueString()
		username = &u
	}
	if !r.Password.IsNull() && !r.Password.IsUnknown() {
		p := r.Password.ValueString()
		password = &p
	}
	return models.RawRecipient{
		Type:     models.RawRecipientType,
		URL:      r.URL.ValueString(),
		Username: username,
		Password: password,
	}
}

func (r *HTTPDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.RawRecipientView); ok {
		r.URL = types.StringValue(v.URL)
		if v.Username != nil {
			r.Username = types.StringValue(*v.Username)
		}
		// Don't modify username if not returned (preserve current state)
		// Password is sensitive and not returned in the view (preserve current state)
	}
}

func (r HTTPDrain) GetDrain() Drain { return r.Drain }

func (r *HTTPDrain) SetDrain(common Drain) { r.Drain = common }

// OVH drain
type OVHDrain struct {
	Drain
	URL                         types.String `tfsdk:"url"`
	Token                       types.String `tfsdk:"token"`
	RFC5424StructuredDataParams types.String `tfsdk:"rfc5424_structured_data_parameters"`
}

func (r OVHDrain) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "OVH logs destination URL",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"token": schema.StringAttribute{
			Required:            true,
			Sensitive:           true,
			MarkdownDescription: "OVH authentication token",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"rfc5424_structured_data_parameters": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "RFC5424 structured data parameters",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r OVHDrain) ToSDKRecipient() models.DrainRecipient {
	var token, params *string
	if !r.Token.IsNull() && !r.Token.IsUnknown() {
		t := r.Token.ValueString()
		token = &t
	}
	if !r.RFC5424StructuredDataParams.IsNull() && !r.RFC5424StructuredDataParams.IsUnknown() {
		p := r.RFC5424StructuredDataParams.ValueString()
		params = &p
	}
	return models.OVHTCPRecipient{
		Type:                            models.OVHTCPRecipientType,
		URL:                             r.URL.ValueString(),
		Token:                           token,
		Rfc5424StructuredDataParameters: params,
	}
}

func (r *OVHDrain) FromSDKRecipient(recipient models.DrainRecipientView) {
	if v, ok := recipient.(models.OVHTCPRecipientView); ok {
		r.URL = types.StringValue(v.URL)
		// Token is sensitive and not returned in the view (preserve current state)
		if v.Rfc5424StructuredDataParameters != nil {
			r.RFC5424StructuredDataParams = types.StringValue(*v.Rfc5424StructuredDataParameters)
		}
		// Don't modify RFC5424 params if not returned (preserve current state)
	}
}

func (r OVHDrain) GetDrain() Drain { return r.Drain }

func (r *OVHDrain) SetDrain(common Drain) { r.Drain = common }
