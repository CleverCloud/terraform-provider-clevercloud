package clevercloud

import (
	"context"
	"strconv"
	"time"

	"github.com/antihax/optional"
	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
		"name_code": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var addonProviderInfoFullResource = &schema.Resource{
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
		"plans": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeSet,
				Elem: addonPlanResource,
			},
		},
		"features": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeSet,
				Elem: addonFeatureResource,
			},
		},
	},
}

func dataSourceAddonProviders() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAddonProvidersRead,
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"addon_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeSet,
					Elem: addonProviderInfoFullResource,
				},
			},
		},
	}
}

func dataSourceAddonProvidersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	var providersList []clevercloud.AddonProviderInfoFullView
	var options clevercloud.GetAddonProvidersOpts

	organizationId, ok := d.GetOk("organization_id")
	if !ok {
		options.OrgaId = optional.EmptyString()
	} else {
		options.OrgaId = optional.NewString(organizationId.(string))
	}

	var err error
	if providersList, _, err = cc.ProductsApi.GetAddonProviders(context.Background(), &options); err != nil {
		return diag.FromErr(err)
	}

	providers := make([]interface{}, 0)
	for _, provider := range providersList {
		providers = append(providers, makeAddonProviderInfoFullResourceSchemaSet(&provider))
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	d.Set("addon_providers", providers)

	return diags
}

func makeAddonFeaturesResourceSchemaList(features []clevercloud.AddonFeatureView) []interface{} {
	list := make([]interface{}, 0)

	for _, feature := range features {
		set := &schema.Set{F: schema.HashResource(addonFeatureInstanceResource)}
		set.Add(map[string]interface{}{
			"name":      feature.Name,
			"type":      feature.Type,
			"name_code": feature.NameCode,
		})
		list = append(list, set)
	}

	return list
}

func makeAddonProviderInfoFullResourceSchemaSet(providerInfo *clevercloud.AddonProviderInfoFullView) *schema.Set {
	set := &schema.Set{F: schema.HashResource(addonProviderInfoFullResource)}

	regions := make([]interface{}, 0)
	for _, region := range providerInfo.Regions {
		regions = append(regions, region)
	}

	plans := make([]interface{}, 0)
	for _, plan := range providerInfo.Plans {
		plans = append(plans, makeAddonPlanResourceSchemaSet(&plan))
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
		"plans":       plans,
		"features":    makeAddonFeaturesResourceSchemaList(providerInfo.Features),
	})

	return set
}
