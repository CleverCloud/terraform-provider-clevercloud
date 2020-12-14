package clevercloud

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/dansmaculotte/clevercloud-go/clevercloud"
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

	if err := d.Set("self", self); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func dataSourceSelf() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSelfRead,
		Schema: map[string]*schema.Schema{
			"self": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"phone": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"city": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"zip_code": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"country": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"avatar": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_date": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"langugage": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"email_validated": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						// "oauth_apps": &schema.Schema{
						// 	Type:     schema.TypeList,
						// 	Computed: true,
						// 	Elem: &schema.Schema{
						// 		Type:     schema.TypeString,
						// 		Computed: true,
						// 	},
						// },
						"admin": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"can_pay": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"preferred_mfa": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"has_password": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
