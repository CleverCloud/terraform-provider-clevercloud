package elasticsearch

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/miton18/helper/set"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

type Elasticsearch struct {
	ID            types.String `tfsdk:"id"`
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the Elasticsearch addon",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
				MarkdownDescription: "Elasticsearch major version (e.g., '7', '8'). Only the major version number is used by the API. Changing this requires replacing the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+$`),
						"version must be a major version number (e.g., '7', '8')",
					),
				},
			},
			"encryption": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable at-rest encryption. This option cannot be changed after creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"kibana": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable Kibana for this Elasticsearch add-on. This option cannot be changed after creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"apm": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable APM (Application Performance Monitoring) for this Elasticsearch add-on. This option cannot be changed after creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
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

func (r *ResourceElasticsearch) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() { // Skip validation when deleting
		return
	}

	plan := helper.From[Elasticsearch](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		infosRes := tmp.GetElasticsearchInfos(ctx, r.Client())
		if infosRes.HasError() {
			res.Diagnostics.AddError("failed to get Elasticsearch provider info", infosRes.Error().Error())
			return
		}
		infos := infosRes.Payload()

		availableVersions := set.New[string]()
		for majorVersion := range infos.Dedicated {
			availableVersions.Add(majorVersion)
		}

		requestedVersion := plan.Version.ValueString()
		if !availableVersions.Contains(requestedVersion) {
			res.Diagnostics.AddAttributeError(
				path.Root("version"),
				"Elasticsearch version not available",
				fmt.Sprintf(
					"version '%s' is not available. Available major versions: %s",
					requestedVersion,
					strings.Join(availableVersions.Slice(), ", "),
				),
			)
		}
	}
}

// The API-to-state mappers for clevercloud_elasticsearch follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromAddon maps the generic add-on view. It is the only source of name, plan
// and region: tmp.Elasticsearch carries none of them.
func (es *Elasticsearch) FromAddon(ctx context.Context, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if es == nil || addon == nil {
		return
	}

	es.Name = pkg.FromStr(addon.Name)
	es.Plan = pkg.FromStr(addon.Plan.Slug)
	es.Region = pkg.FromStr(addon.Region)
}

// FromElasticsearch maps the product view: the toggles, the credentials each one
// unlocks, and the plugin list.
//
// encryption, kibana and apm are Optional + Computed + Default(false), and
// kibana and apm additionally carry RequiresReplace: left null in state after an
// import their defaults create a diff, and the next plan destroys and recreates
// the add-on. They are therefore assigned on every read, absent from the API or
// not (#452).
func (es *Elasticsearch) FromElasticsearch(ctx context.Context, api *tmp.Elasticsearch, diags *diag.Diagnostics) {
	if es == nil || api == nil {
		return
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
}
