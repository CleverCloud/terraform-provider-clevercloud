package clevercloud

import (
	"context"
	"net/http"
	"time"

	"github.com/gaelreyrol/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
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
			"clevercloud_self":             dataSourceSelf(),
			"clevercloud_self_application": dataSourceSelfApplication(),
			"clevercloud_self_addon":       dataSourceSelfAddon(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	secret := d.Get("secret").(string)

	var diags diag.Diagnostics

	client := &http.Client{Timeout: 10 * time.Second}

	if (token != "") && (secret != "") {
		cc := clevercloud.NewClient(&clevercloud.Config{
			Token:  token,
			Secret: secret,
		}, client)

		return cc, diags
	}

	envConfig := clevercloud.GetConfigFromEnv()
	if (envConfig.Token != "") && (envConfig.Secret != "") {
		cc := clevercloud.NewClient(envConfig, client)

		return cc, diags
	}

	userConfig := clevercloud.GetConfigFromUser()
	if (userConfig.Token != "") && (userConfig.Secret != "") {
		cc := clevercloud.NewClient(userConfig, client)

		return cc, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Could not configure Clever Cloud Client",
	})

	return nil, diags
}
