package clevercloud

import (
	"context"
	"fmt"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var addonProviderInfoResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"website": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"short_desc": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"long_desc": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"can_upgrade": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"regions": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

var addonFeatureInstanceResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"value": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computable_value": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name_code": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var addonPlanResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"slug": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"features": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeSet,
				Elem: addonFeatureInstanceResource,
			},
		},
		"price": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"zones": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

func dataSourceAddon() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAddonRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"real_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"provider_info": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     addonProviderInfoResource,
			},
			"plan": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     addonPlanResource,
			},
			"creation_date": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAddonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	var addon clevercloud.AddonView

	organizationId, ok := d.GetOk("organization_id")
	if !ok {
		self, _, err := cc.SelfApi.GetUser(context.Background())
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("organization_id", self.Id)

		if addon, _, err = cc.SelfApi.GetSelfAddonById(context.Background(), d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		var err error
		addon, _, err = cc.OrganisationApi.GetAddonByOrgaAndAddonId(context.Background(), organizationId.(string), d.Get("id").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(addon.Id)

	_ = d.Set("name", addon.Name)
	_ = d.Set("real_id", addon.RealId)
	_ = d.Set("region", addon.Region)
	_ = d.Set("config_keys", addon.ConfigKeys)

	if err := d.Set("provider_info", makeAddonProviderInfoResourceSchemaSet(&addon.Provider)); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set addon provider info bindings (%s): %v", d.Id(), err))
	}

	if err := d.Set("plan", makeAddonPlanResourceSchemaSet(&addon.Plan)); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set addon plan bindings (%s): %v", d.Id(), err))
	}

	return diags
}

func makeAddonPlanResourceSchemaSet(plan *clevercloud.AddonPlanView) *schema.Set {
	set := &schema.Set{F: schema.HashResource(addonPlanResource)}

	planZones := make([]interface{}, len(plan.Zones))
	for i, zone := range plan.Zones {
		planZones[i] = zone
	}

	set.Add(map[string]interface{}{
		"id":       plan.Id,
		"name":     plan.Name,
		"slug":     plan.Slug,
		"price":    float64(plan.Price),
		"zones":    planZones,
		"features": makeAddonFeatureInstancesResourceSchemaList(plan.Features),
	})

	return set
}

func makeAddonFeatureInstancesResourceSchemaList(features []clevercloud.AddonFeatureInstanceView) []interface{} {
	list := make([]interface{}, 0)

	for _, feature := range features {
		set := &schema.Set{F: schema.HashResource(addonFeatureInstanceResource)}
		set.Add(map[string]interface{}{
			"name":             feature.Name,
			"type":             feature.Type,
			"value":            feature.Value,
			"computable_value": feature.ComputableValue,
			"name_code":        feature.NameCode,
		})
		list = append(list, set)
	}

	return list
}

func makeAddonProviderInfoResourceSchemaSet(providerInfo *clevercloud.AddonProviderInfoView) *schema.Set {
	set := &schema.Set{F: schema.HashResource(addonProviderInfoResource)}

	regions := make([]interface{}, 0)
	for _, region := range providerInfo.Regions {
		regions = append(regions, region)
	}

	set.Add(map[string]interface{}{
		"id":          providerInfo.Id,
		"name":        providerInfo.Name,
		"website":     providerInfo.Website,
		"short_desc":  providerInfo.ShortDesc,
		"long_desc":   providerInfo.LongDesc,
		"status":      providerInfo.Status,
		"can_upgrade": providerInfo.CanUpgrade,
		"regions":     regions,
	})

	return set
}
