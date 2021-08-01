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
			"flavors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeSet,
					Elem: applicationFlavorResource,
				},
			},
		},
	}
}

func dataSourceFlavorsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	flavorsList, _, err := cc.ProductsApi.GetFlavors(context.Background(), &clevercloud.GetFlavorsOpts{})
	if err != nil {
		return diag.FromErr(err)
	}

	flavors := make([]interface{}, 0)
	for _, flavor := range flavorsList {
		flavors = append(flavors, makeFlavorResourceSchemaSet(&flavor))
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	d.Set("flavors", flavors)

	return diags
}
