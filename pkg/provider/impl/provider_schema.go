package impl

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

// ProviderData is struct implementation of Provider.GetSchema()
type ProviderData struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	Token        types.String `tfsdk:"token"`
	Secret       types.String `tfsdk:"secret"`
	Organisation types.String `tfsdk:"organisation"`
}

//go:embed provider.md
var providerDoc string

// GetSchema return provider schema
func (p *Provider) Schema(_ context.Context, req provider.SchemaRequest, res *provider.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: providerDoc,
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CleverCloud API endpoint, default to https://api.clever-cloud.com",
			},
			"token": schema.StringAttribute{
				Optional:            true, // can be read from ~/.config by client
				Sensitive:           true,
				MarkdownDescription: "CleverCloud OAuth1 token, can be took from clever-tools after login. This parameter can also be provided via CC_OAUTH_TOKEN environment variable.",
			},
			"secret": schema.StringAttribute{
				Optional:            true, // // can be read from ~/.config by client
				Sensitive:           true,
				MarkdownDescription: "CleverCloud OAuth1 secret, can be took from clever-tools after login. This parameter can also be provided via CC_OAUTH_SECRET environment variable.",
			},
			"organisation": schema.StringAttribute{
				Sensitive:           true,
				Required:            true,
				MarkdownDescription: "CleverCloud organisation, can be either orga_xxx, or user_xxx for personal spaces",
				Validators: []validator.String{
					pkg.NewValidatorRegex("valid owner name", regexp.MustCompile(`^(user|orga)_.{36}`)),
				},
			},
		},
	}
}
