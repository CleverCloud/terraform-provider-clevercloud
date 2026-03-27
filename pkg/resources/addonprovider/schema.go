package addonprovider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addonprovider/validators"
)

//go:embed doc.md
var resourceAddonProviderDoc string

// AddonProvider represents the flat Terraform schema structure
type AddonProvider struct {
	ProviderID        types.String `tfsdk:"provider_id"`
	Name              types.String `tfsdk:"name"`
	ConfigVars        types.Set    `tfsdk:"config_vars"`
	Regions           types.Set    `tfsdk:"regions"`
	Password          types.String `tfsdk:"password"`
	SSOSalt           types.String `tfsdk:"sso_salt"`
	ProductionBaseURL URLValue     `tfsdk:"production_base_url"`
	ProductionSSOURL  URLValue     `tfsdk:"production_sso_url"`
	TestBaseURL       URLValue     `tfsdk:"test_base_url"`
	TestSSOURL        URLValue     `tfsdk:"test_sso_url"`
	Features          []Feature    `tfsdk:"feature"`
	Plans             []Plan       `tfsdk:"plan"`
}

// Feature represents an addon provider feature
// Features define capabilities that can be enabled/configured per plan
type Feature struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

// Plan represents an addon provider plan
// Plans define pricing tiers and their configurations
type Plan struct {
	Name     types.String  `tfsdk:"name"`
	Slug     types.String  `tfsdk:"slug"`
	Price    types.Float64 `tfsdk:"price"`
	ID       types.String  `tfsdk:"id"` // Computed
	Features []PlanFeature `tfsdk:"features"`
}

// PlanFeature represents a feature value within a plan
type PlanFeature struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// Equal checks if two plans are equal (ignoring the ID field which is computed)
func (p Plan) Equal(other Plan) bool {
	// Compare basic fields
	if p.Name.ValueString() != other.Name.ValueString() ||
		p.Slug.ValueString() != other.Slug.ValueString() ||
		p.Price.ValueFloat64() != other.Price.ValueFloat64() {
		return false
	}

	// Compare features count
	if len(p.Features) != len(other.Features) {
		return false
	}

	// Build map of features from p for quick lookup
	pFeatures := make(map[string]string)
	for _, f := range p.Features {
		pFeatures[f.Name.ValueString()] = f.Value.ValueString()
	}

	// Check that all features from other exist in p with same values
	for _, f := range other.Features {
		name := f.Name.ValueString()
		value := f.Value.ValueString()
		if pValue, exists := pFeatures[name]; !exists || pValue != value {
			return false
		}
	}

	return true
}

func (r ResourceAddonProvider) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceAddonProviderDoc,
		Attributes: map[string]schema.Attribute{
			"provider_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier for the addon provider (will be used as manifest ID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the addon provider",
			},
			"config_vars": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Set of environment variables exposed by this addon. Each variable must be prefixed with the provider_id (uppercased, dashes replaced by underscores). Example: for provider_id 'my-service', config vars must start with 'MY_SERVICE_'",
				Validators: []validator.Set{
					validators.ConfigVarsPrefixValidator(),
				},
			},
			"regions": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of regions where this addon provider is available. This is computed from the API and returns all supported regions.",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "API password (minimum 35 characters)",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(35),
				},
			},
			"sso_salt": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "SSO salt (minimum 35 characters)",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(35),
				},
			},
			"production_base_url": schema.StringAttribute{
				CustomType:          URLType{},
				Required:            true,
				MarkdownDescription: "Production base URL for resource provisioning API (must use HTTPS)",
				Validators: []validator.String{
					pkg.HTTPSSchemeValidator(),
				},
			},
			"production_sso_url": schema.StringAttribute{
				CustomType:          URLType{},
				Required:            true,
				MarkdownDescription: "Production SSO login URL (must use HTTPS)",
				Validators: []validator.String{
					pkg.HTTPSSchemeValidator(),
				},
			},

			// Test Environment
			"test_base_url": schema.StringAttribute{
				CustomType:          URLType{},
				Required:            true,
				MarkdownDescription: "Test base URL for resource provisioning API (must use HTTPS)",
				Validators: []validator.String{
					pkg.HTTPSSchemeValidator(),
				},
			},
			"test_sso_url": schema.StringAttribute{
				CustomType:          URLType{},
				Required:            true,
				MarkdownDescription: "Test SSO login URL (must use HTTPS)",
				Validators: []validator.String{
					pkg.HTTPSSchemeValidator(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"feature": schema.SetNestedBlock{
				MarkdownDescription: "Features provided by this addon. Features define capabilities that can be enabled/configured per plan (e.g., disk_size, connection_limit, max_db_size). Each feature has a name and a type that determines how it's represented.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Feature name (e.g., 'disk_size', 'connection_limit', 'max_db_size')",
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Feature type. Valid values: `BOOLEAN`, `INTERVAL`, `FILESIZE`, `NUMBER`, `PERCENTAGE`, `STRING`, `OBJECT`, `BYTES`, `BOOLEAN_SHARED`",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"BOOLEAN",
									"INTERVAL",
									"FILESIZE",
									"NUMBER",
									"PERCENTAGE",
									"STRING",
									"OBJECT",
									"BYTES",
									"BOOLEAN_SHARED",
								),
							},
						},
					},
				},
			},
			"plan": schema.SetNestedBlock{
				MarkdownDescription: "Plans offered by this addon provider. Plans define pricing tiers and service levels (e.g., 'free', 'basic', 'premium'). Each plan has a name, slug (URL-friendly identifier), and price per month in euros.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Plan display name (e.g., 'Free Plan', 'Basic', 'Premium')",
						},
						"slug": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "URL-friendly identifier for the plan (e.g., 'free', 'basic', 'premium'). Must be unique within the provider.",
						},
						"price": schema.Float64Attribute{
							Required:            true,
							MarkdownDescription: "Monthly price in euros (e.g., 0 for free, 9.99 for paid plans)",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the plan (generated by Clever Cloud)",
						},
					},
					Blocks: map[string]schema.Block{
						"features": schema.SetNestedBlock{
							MarkdownDescription: "Feature values for this plan. Each feature must correspond to a feature defined at the provider level. The value format depends on the feature type (e.g., '100' for FILESIZE, 'true' for BOOLEAN, '10' for NUMBER).",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Feature name (must match a feature defined at provider level)",
									},
									"value": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Feature value (format depends on feature type: number for NUMBER/FILESIZE, 'true'/'false' for BOOLEAN, etc.)",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
