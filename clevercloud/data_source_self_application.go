package clevercloud

import (
	"context"
	"net/http"
	"time"

	"github.com/gaelreyrol/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSelfApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := &http.Client{Timeout: 10 * time.Second}

	cc := clevercloud.NewClient(clevercloud.GetConfigFromUser(), client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	selfAPI := clevercloud.NewSelfAPI(cc)

	self, err := selfAPI.GetSelf()
	if err != nil {
		return diag.FromErr(err)
	}

	applications, err := selfAPI.GetApplications(self.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("applications", applications)

	return diags
}

func dataSourceSelfApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSelfRead,
		Schema:      map[string]*schema.Schema{},
	}
}
