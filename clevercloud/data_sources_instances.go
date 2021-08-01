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
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"variant": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     applicationInstanceVariantResource,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"coming_soon": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"max_instances": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tags": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"deployments": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"flavors": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeSet,
								Elem: applicationFlavorResource,
							},
						},
						"default_flavor": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     applicationFlavorResource,
						},
						"build_flavor": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     applicationFlavorResource,
						},
					},
				},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	instancesList, _, err := cc.ProductsApi.GetAvailableInstances(context.Background(), &clevercloud.GetAvailableInstancesOpts{})
	if err != nil {
		return diag.FromErr(err)
	}

	instances := make([]interface{}, 0)
	for _, instance := range instancesList {
		instances = append(instances, map[string]interface{}{
			"name":           instance.Name,
			"description":    instance.Description,
			"type":           instance.Type,
			"version":        instance.Version,
			"variant":        makeInstanceVariantResourceSchemaSet(&instance.Variant),
			"enabled":        instance.Enabled,
			"coming_soon":    instance.ComingSoon,
			"max_instances":  int(instance.MaxInstances),
			"tags":           instance.Tags,
			"deployments":    instance.Deployments,
			"flavors":        makeFlavorsResourceSchemaList(instance.Flavors),
			"default_flavor": makeFlavorResourceSchemaSet(&instance.DefaultFlavor),
			"build_flavor":   makeFlavorResourceSchemaSet(&instance.BuildFlavor),
		})
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	d.Set("instances", instances)

	return diags
}
