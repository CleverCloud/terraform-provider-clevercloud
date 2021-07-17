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

func dataSourceAddonProviders() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAddonProvidersRead,
		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAddonProvidersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	var providers []clevercloud.AddonProviderInfoFullView

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		var err error
		if providers, _, err = cc.ProductsApi.GetAddonProviders(context.Background(), &clevercloud.GetAddonProvidersOpts{
			OrgaId: optional.EmptyString(),
		}); err != nil {
			return diag.FromErr(err)
		}
	} else {
		var err error
		if providers, _, err = cc.ProductsApi.GetAddonProviders(context.Background(), &clevercloud.GetAddonProvidersOpts{
			OrgaId: optional.NewString(organizationID.(string)),
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	providerNames := make([]string, 0)
	for _, provider := range providers {
		providerNames = append(providerNames, provider.Name)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	_ = d.Set("names", providerNames)

	return diags
}
