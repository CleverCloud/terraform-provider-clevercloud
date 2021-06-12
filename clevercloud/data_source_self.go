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
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
	cc := m.(*clevercloud.Client)

	var diags diag.Diagnostics

	selfAPI := clevercloud.NewSelfAPI(cc)

	self, err := selfAPI.GetSelf()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(self.ID)

	_ = d.Set("email", self.Email)
	_ = d.Set("name", self.Name)
	_ = d.Set("phone", self.Phone)
	_ = d.Set("address", self.Address)
	_ = d.Set("city", self.City)
	_ = d.Set("zip_code", self.ZipCode)
	_ = d.Set("country", self.Country)
	_ = d.Set("avatar", self.Avatar)
	_ = d.Set("creation_date", self.CreationDate)
	_ = d.Set("language", self.Language)
	_ = d.Set("email_validated", self.EmailValidated)
	_ = d.Set("oauth_apps", self.OAuthApps)
	_ = d.Set("admin", self.Admin)
	_ = d.Set("can_pay", self.CanPay)
	_ = d.Set("preferred_mfa", self.PreferredMFA)
	_ = d.Set("has_password", self.HasPassword)

	return diags
}
