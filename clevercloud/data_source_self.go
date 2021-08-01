package clevercloud

import (
	"context"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSelf() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSelfRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"phone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"city": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zip_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"country": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_validated": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"oauth_apps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"can_pay": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"preferred_mfa": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"has_password": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceSelfRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	self, _, err := cc.SelfApi.GetUser(context.Background())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(self.Id)

	d.Set("name", self.Name)
	d.Set("email", self.Email)
	d.Set("phone", self.Phone)
	d.Set("address", self.Address)
	d.Set("city", self.City)
	d.Set("zip_code", self.Zipcode)
	d.Set("country", self.Country)
	d.Set("avatar", self.Avatar)
	d.Set("creation_date", self.CreationDate)
	d.Set("language", self.Lang)
	d.Set("email_validated", self.EmailValidated)
	d.Set("oauth_apps", self.OauthApps)
	d.Set("admin", self.Admin)
	d.Set("can_pay", self.CanPay)
	d.Set("preferred_mfa", self.PreferredMFA)
	d.Set("has_password", self.HasPassword)

	return diags
}
