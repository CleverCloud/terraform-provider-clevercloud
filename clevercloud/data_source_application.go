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
			"organization_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.Client)

	var diags diag.Diagnostics

	var application *clevercloud.Application

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		selfAPI := clevercloud.NewSelfAPI(cc)

		self, err := selfAPI.GetSelf()
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("organization_id", self.ID)

		if application, err = selfAPI.GetApplication(self.ID, d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		orgAPI := clevercloud.NewOrganizationAPI(cc)

		org, err := orgAPI.GetOrganization(organizationID.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		if application, err = orgAPI.GetApplication(org.ID, d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(application.ID)

	_ = d.Set("name", application.Name)

	return diags
}
