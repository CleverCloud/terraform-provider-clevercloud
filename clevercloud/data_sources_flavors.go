package clevercloud

import (
	"context"
	"strconv"
	"time"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFlavors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlavorsRead,
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

func dataSourceFlavorsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	flavors, _, err := cc.ProductsApi.GetFlavors(context.Background(), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	flavorNames := make([]string, 0)
	for _, flavor := range flavors {
		flavorNames = append(flavorNames, flavor.Name)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	_ = d.Set("names", flavorNames)

	return diags
}
