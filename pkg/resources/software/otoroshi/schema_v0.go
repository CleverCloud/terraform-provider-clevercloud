package otoroshi

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OtoroshiV0 is the state layout of version 0, where networkgroups attached the
// add-on to existing networkgroups through its backing Java application
type OtoroshiV0 struct {
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

	Networkgroups types.Set `tfsdk:"networkgroups"`
}

var schemaOtoroshiV0 = schema.Schema{
	Version: 0,
	Attributes: map[string]schema.Attribute{
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
		// Frozen copy of the shared attribute as it was published, so that later
		// changes to it cannot alter how a version 0 state is decoded
		"networkgroups": schema.SetNestedAttribute{
			Optional:            true,
			MarkdownDescription: "List of networkgroups the application must be part of",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"networkgroup_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "ID of the networkgroup",
					},
					"fqdn": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "domain name which will resolve to application instances inside the networkgroup",
					},
				},
			},
		},
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
