package otoroshi

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Otoroshi struct {
	ID types.String `tfsdk:"id"`

	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`

	CreationDate types.Int64  `tfsdk:"creation_date"`
	Version      types.String `tfsdk:"version"`

	APIClientID          types.String `tfsdk:"api_client_id"`
	APIClientSecret      types.String `tfsdk:"api_client_secret"`
	APIURL               types.String `tfsdk:"api_url"`
	InitialAdminLogin    types.String `tfsdk:"initial_admin_login"`
	InitialAdminPassword types.String `tfsdk:"initial_admin_password"`
	URL                  types.String `tfsdk:"url"`
}

//go:embed doc.md
var resourceOtoroshiDoc string

func (r ResourceOtoroshi) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceOtoroshiDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Otoroshi",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Geographical region where the data will be stored",
				Default:             stringdefault.StaticString("par"),
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Otoroshi version to deploy",
			},

			// provider
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated unique identifier",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"creation_date": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Date of Otoroshi creation",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"api_client_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "API client ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"api_client_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "API client secret",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"api_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "API URL",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"initial_admin_login": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Initial admin login",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"initial_admin_password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial admin password",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
