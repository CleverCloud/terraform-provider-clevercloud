package clevercloud

import (
	"context"
	"strconv"
	"time"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZonesRead,
		Schema: map[string]*schema.Schema{
			"zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"internal": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"corresponding_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	zonesList, _, err := cc.ProductsApi.GetZones(context.Background())
	if err != nil {
		return diag.FromErr(err)
	}

	zones := make([]interface{}, 0)
	for _, zone := range zonesList {
		zones = append(zones, map[string]interface{}{
			"name":                 zone.Name,
			"internal":             zone.Internal,
			"corresponding_region": zone.CorrespondingRegion,
		})
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	d.Set("zones", zones)

	return diags
}
