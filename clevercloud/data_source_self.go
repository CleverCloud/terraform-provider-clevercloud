package clevercloud

import (
	"context"
	"net/http"
	"time"

	"github.com/gaelreyrol/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSelfRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := &http.Client{Timeout: 10 * time.Second}

	cc := clevercloud.NewClient(clevercloud.GetConfigFromUser(), client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	selfAPI := clevercloud.NewSelfAPI(cc)

	self, err := selfAPI.GetSelf()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(self.ID)

	d.Set("email", self.Email)
	d.Set("name", self.Name)
	d.Set("phone", self.Phone)
	d.Set("address", self.Address)
	d.Set("city", self.City)
	d.Set("zip_code", self.ZipCode)
	d.Set("country", self.Country)
	d.Set("avatar", self.Avatar)
	d.Set("creation_date", self.CreationDate)
	d.Set("language", self.Language)
	d.Set("email_validated", self.EmailValidated)
	d.Set("oauth_apps", self.OAuthApps)
	d.Set("admin", self.Admin)
	d.Set("can_pay", self.CanPay)
	d.Set("preferred_mfa", self.PreferredMFA)
	d.Set("has_password", self.HasPassword)

	return diags
}

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
