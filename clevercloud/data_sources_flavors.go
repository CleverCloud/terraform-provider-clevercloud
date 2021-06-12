package clevercloud

import (
	"context"
	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceFlavors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlavorsRead,
		Schema: map[string]*schema.Schema{
			"names": &schema.Schema{
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
	cc := m.(*clevercloud.Client)

	var diags diag.Diagnostics

	productAPI := clevercloud.NewProductAPI(cc)

	flavors, err := productAPI.GetFlavors()
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
