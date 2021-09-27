package clevercloud

import (
	"context"

	"github.com/clevercloud/clevercloud-go/clevercloud"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	CONSUMER_KEY    = "CHANGEME"
	CONSUMER_SECRET = "CHANGEME"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *clevercloud.APIClient
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Type:     types.StringType,
				Required: true,
			},
			"secret": {
				Type:      types.StringType,
				Required:  true,
				Sensitive: true,
			},
			"consumer_key": {
				Type:     types.StringType,
				Optional: true,
			},
			"consumer_secret": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
		},
	}, nil
}

type providerData struct {
	Token          types.String `tfsdk:"token"`
	Secret         types.String `tfsdk:"secret"`
	ConsumerKey    types.String `tfsdk:"consumer_key"`
	ConsumerSecret types.String `tfsdk:"consumer_secret"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData

	if diags := req.Config.Get(ctx, &config); diags != nil {
		resp.Diagnostics.AddError("Error parsing Clever Cloud configuration", "Error parsing the configuration, this is an error in the provider.")
		return
	}

	consumerKey := CONSUMER_KEY
	consumerSecret := CONSUMER_SECRET

	if !config.ConsumerKey.Unknown {
		consumerKey = config.ConsumerKey.Value
	}

	if !config.ConsumerSecret.Unknown {
		consumerSecret = config.ConsumerSecret.Value
	}

	if config.Token.Unknown || config.Secret.Unknown {
		resp.Diagnostics.AddError("Unable to create Clever Cloud client", "Config token or secret are missing.")
		return
	}

	oauth := clevercloud.NewOAuthClient(consumerKey, consumerSecret)
	oauth.SetTokens(config.Token.Value, config.Secret.Value)

	api := clevercloud.NewOAuthAPIClient(oauth, clevercloud.NewConfiguration())

	p.client = api
	p.configured = true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"clevercloud_application": resourceApplicationType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
