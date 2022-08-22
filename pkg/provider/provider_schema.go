package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

const providerDoc = `
CleverCloud provider allow you to interract with CleverCloud platform.
`

// GetSchema return provider schema
func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: providerDoc,
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CleverCloud API endpoint, default to https://api.clever-cloud.com",
			},
			"token": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CleverCloud OAuth1 token, can be took from clever-tools after login",
			},
			"secret": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "CleverCloud OAuth1 secret, can be took from clever-tools after login",
			},
			"organisation": {
				Type:                types.StringType,
				Sensitive:           true,
				Required:            true,
				MarkdownDescription: "CleverCloud organisation, can be either orga_xxx, or user_xxx for personal spaces",
				Validators: []tfsdk.AttributeValidator{
					pkg.NewValidatorRegex("valid owner name", regexp.MustCompile(`^(user|orga)_.{36}`)),
				},
			},
		},
	}, nil
}
