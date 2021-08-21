package clevercloud

import (
	"context"

	"github.com/clevercloud/clevercloud-go/clevercloud"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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

func (p *provider) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	err := req.Config.Get(ctx, &config)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error parsing Clever Cloud configuration",
			Detail:   "Error parsing the configuration, this is an error in the provider:\n\n" + err.Error(),
		})
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
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Unable to create Clever Cloud client",
		})
		return
	}

	oauth := clevercloud.NewOAuthClient(consumerKey, consumerSecret)
	oauth.SetTokens(config.Token.Value, config.Secret.Value)

	api := clevercloud.NewOAuthAPIClient(oauth, clevercloud.NewConfiguration())

	p.client = api
	p.configured = true
}

func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, []*tfprotov6.Diagnostic) {
	return map[string]tfsdk.ResourceType{
		"clevercloud_application": resourceApplicationType{},
	}, nil
}

func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, []*tfprotov6.Diagnostic) {
	return map[string]tfsdk.DataSourceType{}, nil
}
