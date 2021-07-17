package clevercloud

import (
	"context"
	"strconv"
	"time"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstancesRead,
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

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	instances, _, err := cc.ProductsApi.GetAvailableInstances(context.Background(), &clevercloud.GetAvailableInstancesOpts{})
	if err != nil {
		return diag.FromErr(err)
	}

	instanceNames := make([]string, 0)
	for _, instance := range instances {
		instanceNames = append(instanceNames, instance.Name)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	_ = d.Set("names", instanceNames)

	return diags
}
