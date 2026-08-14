package route

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Route struct {
	ID types.String `tfsdk:"id"`

	OtoroshiID    types.String `tfsdk:"otoroshi_id"`
	Name          types.String `tfsdk:"name"`
	Domain        types.String `tfsdk:"domain"`
	Path          types.String `tfsdk:"path"`
	StripPath     types.Bool   `tfsdk:"strip_path"`
	ApplicationID types.String `tfsdk:"application_id"`
	Port          types.Int64  `tfsdk:"port"`
	Enabled       types.Bool   `tfsdk:"enabled"`

	BackendHostname types.String `tfsdk:"backend_hostname"`
	URL             types.String `tfsdk:"url"`
}

//go:embed doc.md
var resourceRouteDoc string

func (r ResourceOtoroshiRoute) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceRouteDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"otoroshi_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the Otoroshi add-on serving this route",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the route, as displayed in the Otoroshi console",
			},
			"application_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the Clever Cloud application the route forwards to. It must be a member of the Otoroshi networkgroup",
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Port the application listens on inside the networkgroup",
			},
			"domain": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Public domain the route answers on. Defaults to the `cleverapps.io` domain of the Otoroshi entrypoint. Any domain attached to that application works, unlike the `services.clever-cloud.com` ones which are reserved for Otoroshi's own endpoints",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path prefix the route answers on, such as `/my-app`. When omitted the route answers on the whole domain",
			},
			"strip_path": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Remove `path` from the request before forwarding it to the application",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable or disable the route without deleting it",
			},

			// provider
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated unique identifier",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"backend_hostname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Networkgroup FQDN the route forwards to, derived from the application and the Otoroshi networkgroup",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Public URL the route answers on",
			},
		},
	}
}
