package drain

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
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
	ToRecipient() []byte
	FromAPI(drain tmp.Drain) error
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

func (r DatadogDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientDatadog{
		Type: "DatadogRecipient",
		URL:  r.URL.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *DatadogDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientDatadog
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
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

func (r NewRelicDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientNewrelic{
		Type:   "NewRelicRecipient",
		URL:    r.URL.ValueString(),
		APIKey: r.APIKey.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *NewRelicDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientNewrelic
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	// For sensitive attributes, preserve current state value since API doesn't return it
	if !r.APIKey.IsUnknown() {
		// Keep existing value
	} else {
		r.APIKey = types.StringValue(recipient.APIKey)
	}
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
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

func (r ElasticsearchDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientElasticsearch{
		Type:            "ElasticsearchRecipient",
		URL:             r.URL.ValueString(),
		Username:        r.Username.ValueString(),
		Password:        r.Password.ValueString(),
		Index:           r.Index.ValueString(),
		TLSVerification: r.TLSVerification.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *ElasticsearchDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientElasticsearch
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	r.Username = types.StringValue(recipient.Username)
	// For sensitive attributes, preserve current state value since API doesn't return it
	if !r.Password.IsUnknown() {
		// Keep existing value
	} else {
		r.Password = types.StringValue(recipient.Password)
	}
	r.Index = types.StringValue(recipient.Index)
	r.TLSVerification = types.StringValue(recipient.TLSVerification)
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
}

func (r ElasticsearchDrain) GetDrain() Drain { return r.Drain }

func (r *ElasticsearchDrain) SetDrain(drain Drain) { r.Drain = drain }

// Syslog UDP drain
type SyslogUDPDrain struct {
	Drain
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
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
		"token": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Authentication token",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r SyslogUDPDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientSyslogUDP{
		Type:  "SyslogUDPRecipient",
		URL:   r.URL.ValueString(),
		Token: r.Token.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *SyslogUDPDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientSyslogUDP
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	// For sensitive attributes, preserve current state value since API doesn't return it
	if !r.Token.IsUnknown() {
		// Keep existing value
	} else {
		if recipient.Token != "" {
			r.Token = types.StringValue(recipient.Token)
		} else {
			r.Token = types.StringNull()
		}
	}
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
}

func (r SyslogUDPDrain) GetDrain() Drain { return r.Drain }

func (r *SyslogUDPDrain) SetDrain(common Drain) { r.Drain = common }

// Syslog TCP drain
type SyslogTCPDrain struct {
	Drain
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
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
		"token": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Authentication token",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r SyslogTCPDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientSyslogTCP{
		Type:  "SyslogTCPRecipient",
		URL:   r.URL.ValueString(),
		Token: r.Token.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *SyslogTCPDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientSyslogTCP
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	// For sensitive attributes, preserve current state value since API doesn't return it
	if !r.Token.IsUnknown() {
		// Keep existing value
	} else {
		if recipient.Token != "" {
			r.Token = types.StringValue(recipient.Token)
		} else {
			r.Token = types.StringNull()
		}
	}
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
}

func (r SyslogTCPDrain) GetDrain() Drain { return r.Drain }

func (r *SyslogTCPDrain) SetDrain(common Drain) { r.Drain = common }

// HTTP drain
type HTTPDrain struct {
	Drain
	URL types.String `tfsdk:"url"`
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
	}
}

func (r HTTPDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientRaw{
		Type: "RawRecipient",
		URL:  r.URL.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *HTTPDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientRaw
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
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

func (r OVHDrain) ToRecipient() []byte {
	v, err := json.Marshal(tmp.RecipientOVH{
		Type:                        "OVHTCPRecipient",
		URL:                         r.URL.ValueString(),
		Token:                       r.Token.ValueString(),
		RFC5424StructuredDataParams: r.RFC5424StructuredDataParams.ValueString(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *OVHDrain) FromAPI(drain tmp.Drain) error {
	var recipient tmp.RecipientOVH
	if err := json.Unmarshal(drain.Recipient, &recipient); err != nil {
		return err
	}
	r.URL = types.StringValue(recipient.URL)
	// For sensitive attributes, preserve current state value since API doesn't return it
	if !r.Token.IsUnknown() {
		// Keep existing value
	} else {
		r.Token = types.StringValue(recipient.Token)
	}
	if recipient.RFC5424StructuredDataParams != "" {
		r.RFC5424StructuredDataParams = types.StringValue(recipient.RFC5424StructuredDataParams)
	} else {
		r.RFC5424StructuredDataParams = types.StringNull()
	}
	r.ID = types.StringValue(drain.ID)
	r.ResourceID = types.StringValue(drain.ApplicationID)
	r.Kind = types.StringValue(string(drain.Kind))
	return nil
}

func (r OVHDrain) GetDrain() Drain { return r.Drain }

func (r *OVHDrain) SetDrain(common Drain) { r.Drain = common }
