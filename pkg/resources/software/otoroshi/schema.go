package otoroshi

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type Otoroshi struct {
	addon.CommonAttributes

	Version types.String `tfsdk:"version"`

	APIClientID          types.String `tfsdk:"api_client_id"`
	APIClientSecret      types.String `tfsdk:"api_client_secret"`
	APIURL               types.String `tfsdk:"api_url"`
	InitialAdminLogin    types.String `tfsdk:"initial_admin_login"`
	InitialAdminPassword types.String `tfsdk:"initial_admin_password"`
	URL                  types.String `tfsdk:"url"`
}

func (o *Otoroshi) GetCommonPtr() *addon.CommonAttributes {
	return &o.CommonAttributes
}

func (o *Otoroshi) GetAddonOptions() map[string]string {
	opts := map[string]string{}
	if !o.Version.IsNull() && !o.Version.IsUnknown() {
		opts["version"] = o.Version.ValueString()
	}
	return opts
}

func (o *Otoroshi) SetFromResponse(ctx context.Context, cc *client.Client, org string, addonID string, diags *diag.Diagnostics) {
	otoroshiRes := tmp.GetOtoroshi(ctx, cc, o.ID.ValueString())
	if otoroshiRes.HasError() {
		diags.AddError("failed to get Otoroshi", otoroshiRes.Error().Error())
		return
	}
	otoroshi := otoroshiRes.Payload()
	o.APIURL = pkg.FromStr(otoroshi.API.URL)
	o.APIClientID = pkg.FromStr(otoroshi.EnvVars["CC_OTOROSHI_API_CLIENT_ID"])
	o.APIClientSecret = pkg.FromStr(otoroshi.EnvVars["CC_OTOROSHI_API_CLIENT_SECRET"])
	o.InitialAdminLogin = pkg.FromStr(otoroshi.Initialredentials.User)
	o.InitialAdminPassword = pkg.FromStr(otoroshi.Initialredentials.Passsword)
	o.URL = pkg.FromStr(otoroshi.AccessURL)
}

func (o *Otoroshi) SetDefaults() {}

//go:embed doc.md
var resourceOtoroshiDoc string

func (r ResourceOtoroshi) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceOtoroshiDoc,
		Attributes: addon.WithAddonCommons(map[string]schema.Attribute{
			// Single-plan addon: plan is computed, not user-specified
			"plan": schema.StringAttribute{Computed: true},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Otoroshi version to deploy",
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
		}),
	}
}
