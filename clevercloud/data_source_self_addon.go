package clevercloud

import (
	"context"

	"github.com/gaelreyrol/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSelfAddon() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSelfAddonRead,
		Schema:      map[string]*schema.Schema{},
	}
}

func dataSourceSelfAddonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.Client)

	var diags diag.Diagnostics

	selfAPI := clevercloud.NewSelfAPI(cc)

	self, err := selfAPI.GetSelf()
	if err != nil {
		return diag.FromErr(err)
	}

	addons, err := selfAPI.GetAddons(self.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("addons", addons)

	return diags
}
