package clevercloud

import (
	"context"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": &schema.Schema{
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

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	var application clevercloud.ApplicationView

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		self, _, err := cc.SelfApi.GetUser(context.Background())
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("organization_id", self.Id)

		if application, _, err = cc.SelfApi.GetSelfApplicationByAppId(context.Background(), d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		var err error
		if application, _, err = cc.OrganisationApi.GetApplicationByOrgaAndAppId(context.Background(), organizationID.(string), d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(application.Id)

	_ = d.Set("name", application.Name)
	_ = d.Set("state", application.State)
	_ = d.Set("zone", application.Zone)

	return diags
}
