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
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	zones, _, err := cc.ProductsApi.GetZones(context.Background())
	if err != nil {
		return diag.FromErr(err)
	}

	zoneNames := make([]string, 0)
	for _, zone := range zones {
		zoneNames = append(zoneNames, zone.Name)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	_ = d.Set("names", zoneNames)

	return diags
}
