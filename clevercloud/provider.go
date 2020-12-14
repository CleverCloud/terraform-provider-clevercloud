package clevercloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"clevercloud_self":             dataSourceSelf(),
			"clevercloud_self_application": dataSourceSelfApplication(),
			"clevercloud_self_addon":       dataSourceSelfAddon(),
		},
	}
}
