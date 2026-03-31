package oauth_consumer

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OAuthConsumer struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	BaseURL     types.String `tfsdk:"base_url"`
	LogoURL     types.String `tfsdk:"logo_url"`
	WebsiteURL  types.String `tfsdk:"website_url"`
	Rights      types.Set    `tfsdk:"rights"`
	Secret      types.String `tfsdk:"secret"`
}

//go:embed doc.md
var resourceOAuthConsumerDoc string

func (r ResourceOAuthConsumer) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceOAuthConsumerDoc,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OAuth consumer unique identifier (also used as the OAuth client ID/key)",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the OAuth consumer",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the OAuth consumer",
			},
			"base_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Base URL of the OAuth consumer application",
			},
			"logo_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL of the consumer's logo/picture",
			},
			"website_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Website URL of the OAuth consumer",
			},
			"rights": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Set of OAuth rights/permissions. Available rights: access_organisations, access_organisations_bills, access_organisations_consumption_statistics, access_organisations_credit_count, access_personal_information, manage_organisations, manage_organisations_applications, manage_organisations_members, manage_organisations_services, manage_personal_information, manage_ssh_keys",
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
				Validators:          []validator.Set{ValidateRights()},
			},
			"secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "OAuth consumer secret (client secret)",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
