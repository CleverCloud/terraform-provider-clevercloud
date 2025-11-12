package elasticsearch

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Elasticsearch struct {
	Name          types.String `tfsdk:"name"`
	Plan          types.String `tfsdk:"plan"`
	Region        types.String `tfsdk:"region"`
	Networkgroups types.Set    `tfsdk:"networkgroups"`

	Version    types.String `tfsdk:"version"`
	Encryption types.Bool   `tfsdk:"encryption"`
	Kibana     types.Bool   `tfsdk:"kibana"`
	Apm        types.Bool   `tfsdk:"apm"`
	Plugins    types.Set    `tfsdk:"plugins"`

	Host           types.String `tfsdk:"host"`
	User           types.String `tfsdk:"user"`
	Password       types.String `tfsdk:"password"`
	KibanaHost     types.String `tfsdk:"kibana_host"`
	KibanaUser     types.String `tfsdk:"kibana_user"`
	KibanaPassword types.String `tfsdk:"kibana_password"`
	ApmHost        types.String `tfsdk:"apm_host"`
	ApmUser        types.String `tfsdk:"apm_user"`
	ApmPassword    types.String `tfsdk:"apm_password"`
	ApmToken       types.String `tfsdk:"apm_token"`
}

//go:embed doc.md
var resourceElasticsearchDoc string

func (r *ResourceElasticsearch) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceElasticsearchDoc,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Name of the elasticsearch"},
			"plan": schema.StringAttribute{Required: true, MarkdownDescription: "Database size and spec"},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("par"),
				MarkdownDescription: "Geographical region where the data will be stored",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Elasticsearch version",
				Validators: []validator.String{
					pkg.NewStringValidator("Match existing Elasticsearch version", r.validateESVersion),
				},
			},
			"encryption": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable at-rest encryption",
			},
			"kibana": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable Kibana for this Elasticsearch add-on.",
			},
			"apm": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable APM (Application Performance Monitoring) for this Elasticsearch add-on.",
			},
			"plugins": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of plugins to install",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"host":            schema.StringAttribute{Computed: true, MarkdownDescription: "Elasticsearch host, used to connect to"},
			"user":            schema.StringAttribute{Computed: true, MarkdownDescription: "Login username"},
			"password":        schema.StringAttribute{Computed: true, MarkdownDescription: "Login password", Sensitive: true},
			"kibana_host":     schema.StringAttribute{Computed: true, MarkdownDescription: "Kibana URL if enabled"},
			"kibana_user":     schema.StringAttribute{Computed: true, MarkdownDescription: "Kibana user"},
			"kibana_password": schema.StringAttribute{Computed: true, MarkdownDescription: "Kibana password", Sensitive: true},
			"apm_host":        schema.StringAttribute{Computed: true, MarkdownDescription: "APM URL if enabled"},
			"apm_user":        schema.StringAttribute{Computed: true, MarkdownDescription: "APM user"},
			"apm_password":    schema.StringAttribute{Computed: true, MarkdownDescription: "APM password", Sensitive: true},
			"apm_token":       schema.StringAttribute{Computed: true, MarkdownDescription: "APM token", Sensitive: true},
			"networkgroups": schema.SetNestedAttribute{
				Optional:            true,
				MarkdownDescription: "List of networkgroups the addon must be part of",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"networkgroup_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "ID of the networkgroup",
						},
						"fqdn": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "domain name which will resolve to addon instances inside the networkgroup",
						},
					},
				},
			},
		},
	}
}

func (r *ResourceElasticsearch) validateESVersion(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	// TODO
	/*version := req.ConfigValue.ValueString()
	//for v := range r.versions.Iter() {

	fmt.Printf("\n\nversion: %+v\n\n", r.versions)
	//}
	if req.ConfigValue.IsNull() || !r.versions.Contains(version) {
		versions := r.versions.Slice()
		res.Diagnostics.AddError(
			"invalid Elastic version",
			fmt.Sprintf("available versions are: %s", strings.Join(versions, ", ")),
		)
	}*/
}
