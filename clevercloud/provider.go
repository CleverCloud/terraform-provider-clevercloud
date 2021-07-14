package clevercloud

import (
	"context"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	CONSUMER_KEY    = "CHANGEME"
	CONSUMER_SECRET = "CHANGEME"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"secret": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"consumer_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"consumer_secret": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"clevercloud_self":        dataSourceSelf(),
			"clevercloud_application": dataSourceApplication(),
			"clevercloud_addon":       dataSourceAddon(),
			"clevercloud_zones":       dataSourceZones(),
			"clevercloud_flavors":     dataSourceFlavors(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	secret := d.Get("secret").(string)

	consumerKey := CONSUMER_KEY
	consumerSecret := CONSUMER_SECRET

	if d.Get("consumer_key").(string) != "" && d.Get("consumer_secret").(string) != "" {
		consumerKey = d.Get("consumer_key").(string)
		consumerSecret = d.Get("consumer_secret").(string)
	}

	var diags diag.Diagnostics

	if (token != "") && (secret != "") {
		oauth := clevercloud.NewOAuthClient(consumerKey, consumerSecret)
		oauth.SetTokens(token, secret)

		config := clevercloud.NewConfiguration()
		api := clevercloud.NewOAuthAPIClient(oauth, config)

		return api, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Could not configure Clever Cloud Client",
	})

	return nil, diags
}
