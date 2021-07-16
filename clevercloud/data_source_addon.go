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
		"id": &schema.Schema{
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

var addonFeatureResource = &schema.Resource{
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
		"id": &schema.Schema{
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
				Elem: addonFeatureResource,
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
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"real_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_keys": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"provider_info": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     addonProviderInfoResource,
			},
			"plan": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     addonPlanResource,
			},
			"creation_date": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organization_id": &schema.Schema{
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

	organizationID, ok := d.GetOk("organization_id")
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
		addon, _, err = cc.OrganisationApi.GetAddonByOrgaAndAddonId(context.Background(), organizationID.(string), d.Get("id").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(addon.Id)

	_ = d.Set("name", addon.Name)
	_ = d.Set("real_id", addon.RealId)
	_ = d.Set("region", addon.Region)
	_ = d.Set("config_keys", addon.ConfigKeys)

	providerInfoRegions := make([]interface{}, len(addon.Provider.Regions))
	for i, region := range addon.Provider.Regions {
		providerInfoRegions[i] = region
	}

	providerInfoBindings := &schema.Set{F: schema.HashResource(addonProviderInfoResource)}
	providerInfoBindings.Add(map[string]interface{}{
		"id":          addon.Provider.Id,
		"name":        addon.Provider.Name,
		"website":     addon.Provider.Website,
		"short_desc":  addon.Provider.ShortDesc,
		"long_desc":   addon.Provider.LongDesc,
		"status":      addon.Provider.Status,
		"can_upgrade": addon.Provider.CanUpgrade,
		"regions":     providerInfoRegions,
	})

	if err := d.Set("provider_info", providerInfoBindings); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set addon provider info bindings (%s): %v", d.Id(), err))
	}

	planZones := make([]interface{}, len(addon.Plan.Zones))
	for i, zone := range addon.Plan.Zones {
		planZones[i] = zone
	}

	planFeatures := make([]interface{}, len(addon.Plan.Features))
	for i, feature := range addon.Plan.Features {
		planFeature := &schema.Set{F: schema.HashResource(addonFeatureResource)}
		planFeature.Add(map[string]interface{}{
			"name":             feature.Name,
			"type":             feature.Type,
			"value":            feature.Value,
			"computable_value": feature.ComputableValue,
			"name_code":        feature.NameCode,
		})
		planFeatures[i] = planFeature
	}

	planBindings := &schema.Set{F: schema.HashResource(addonPlanResource)}
	planBindings.Add(map[string]interface{}{
		"id":       addon.Plan.Id,
		"name":     addon.Plan.Name,
		"slug":     addon.Plan.Slug,
		"price":    float64(addon.Plan.Price),
		"zones":    planZones,
		"features": planFeatures,
	})

	if err := d.Set("plan", planBindings); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set addon plan bindings (%s): %v", d.Id(), err))
	}

	return diags
}
