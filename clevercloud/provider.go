package clevercloud

import (
	"context"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	consumerKey    = "Q0BrzJQ44MBbWN8cMEkSA6zTBtu2jR"
	consumerSecret = "Fc18B5dIrWsdUEh881WGvo1DxjKC6p"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret": &schema.Schema{
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
