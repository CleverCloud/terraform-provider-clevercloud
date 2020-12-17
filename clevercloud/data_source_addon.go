package clevercloud

import (
	"context"

	"github.com/gaelreyrol/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAddon() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAddonRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAddonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.Client)

	var diags diag.Diagnostics

	var addon *clevercloud.Addon

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		selfAPI := clevercloud.NewSelfAPI(cc)

		self, err := selfAPI.GetSelf()
		if err != nil {
			return diag.FromErr(err)
		}

		if addon, err = selfAPI.GetAddon(self.ID, d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		orgAPI := clevercloud.NewOrganizationAPI(cc)

		org, err := orgAPI.GetOrganization(organizationID.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		if addon, err = orgAPI.GetAddon(org.ID, d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(addon.ID)

	_ = d.Set("name", addon.Name)

	return diags
}
