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

// GetSchema return provider schema
func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"token": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"secret": {
				Type:      types.StringType,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"organisation": {
				Type:      types.StringType,
				Sensitive: true,
				Required:  true,
				Validators: []tfsdk.AttributeValidator{
					pkg.NewValidatorRegex("valid owner name", regexp.MustCompile(`^(user|org)_.{36}`)),
				},
			},
		},
	}, nil
}
