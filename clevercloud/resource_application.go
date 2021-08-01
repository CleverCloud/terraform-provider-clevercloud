package clevercloud

import (
	"context"
	"sort"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"zone": {
				Type:    schema.TypeString,
				Default: "par",
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"min_instances": {
				Type:    schema.TypeInt,
				Default: 1,
			},
			"max_instances": {
				Type:    schema.TypeInt,
				Default: 1,
			},
			"min_flavor": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_flavor": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"archived": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"homogeneous": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sticky_sessions": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cancel_on_push": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"force_https": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"favorite": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"shutdownable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"separate_build": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"build_flavor": {
				Type:         schema.TypeString,
				RequiredWith: []string{"separate_build"},
			},
		},
	}
}

func getInstanceByType(cc *clevercloud.APIClient, instanceType string) (*clevercloud.AvailableInstanceView, diag.Diagnostics) {
	instances, _, err := cc.ProductsApi.GetAvailableInstances(context.Background(), &clevercloud.GetAvailableInstancesOpts{})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	enabledInstances := make([]clevercloud.AvailableInstanceView, 0)
	for _, instance := range instances {
		if instance.Enabled {
			enabledInstances = append(enabledInstances, instance)
		}
	}

	matchingInstances := make([]clevercloud.AvailableInstanceView, 0)
	for _, instance := range enabledInstances {
		if instance.Variant.Slug == instanceType {
			matchingInstances = append(matchingInstances, instance)
		}
	}

	instanceVersions := make([]string, 0)
	for _, instance := range matchingInstances {
		instanceVersions = append(instanceVersions, instance.Version)
	}

	sort.Strings(instanceVersions)

	latestInstanceVersion := instanceVersions[len(instanceVersions)-1]

	var latestInstance clevercloud.AvailableInstanceView
	for _, instance := range matchingInstances {
		if instance.Version == latestInstanceVersion {
			latestInstance = instance
		}
	}

	return &latestInstance, nil
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	description, ok := d.GetOk("description")
	if !ok {
		description = d.Get("name")
	}

	instance, errors := getInstanceByType(cc, d.Get("type").(string))
	if errors != nil {
		diags = append(diags, errors...)
	}

	defaultFlavorName := instance.DefaultFlavor.Name

	wannabeApplication := clevercloud.WannabeApplication{
		Name:            d.Get("name").(string),
		Description:     description.(string),
		Zone:            d.Get("zones").(string),
		InstanceType:    instance.Type,
		InstanceVersion: instance.Version,
		InstanceVariant: instance.Variant.Id,
		MinInstances:    d.Get("min_instances").(int32),
		MaxInstances:    d.Get("max_instances").(int32),
		MinFlavor:       defaultFlavorName,
		MaxFlavor:       defaultFlavorName,
	}

	for _, flavor := range instance.Flavors {
		minFlavorName, ok := d.GetOk("min_flavor")
		if ok && minFlavorName.(string) == flavor.Name {
			wannabeApplication.MinFlavor = minFlavorName.(string)
		}
	}

	for _, flavor := range instance.Flavors {
		maxFlavorName, ok := d.GetOk("max_flavor")
		if ok && maxFlavorName.(string) == flavor.Name {
			wannabeApplication.MaxFlavor = maxFlavorName.(string)
		}
	}

	if tags, ok := d.GetOk("tags"); ok {
		wannabeApplication.Tags = tags.([]string)
	}

	if homogeneous, ok := d.GetOkExists("homogeneous"); ok {
		wannabeApplication.Homogeneous = homogeneous.(bool)
	}

	if stickySessions, ok := d.GetOkExists("sticky_sessions"); ok {
		wannabeApplication.StickySessions = stickySessions.(bool)
	}

	if cancelOnPush, ok := d.GetOkExists("cancel_on_push"); ok {
		wannabeApplication.CancelOnPush = cancelOnPush.(bool)
	}

	if forceHTTPS, ok := d.GetOkExists("force_https"); ok {
		if forceHTTPS.(bool) {
			wannabeApplication.ForceHttps = "ENABLED"
		} else {
			wannabeApplication.ForceHttps = "DISABLED"
		}
	}

	if favorite, ok := d.GetOkExists("favorite"); ok {
		wannabeApplication.Favourite = favorite.(bool)
	}

	if shutdownable, ok := d.GetOkExists("shutdownable"); ok {
		wannabeApplication.Shutdownable = shutdownable.(bool)
	}

	if separateBuild, ok := d.GetOkExists("separate_build"); ok {
		wannabeApplication.SeparateBuild = separateBuild.(bool)
	}

	if wannabeApplication.SeparateBuild {
		buildFlavorName, ok := d.GetOk("build_flavor")
		if ok {
			for _, flavor := range instance.Flavors {
				if buildFlavorName.(string) == flavor.Name {
					wannabeApplication.BuildFlavor = buildFlavorName.(string)
				}
			}
		} else {
			wannabeApplication.BuildFlavor = defaultFlavorName
		}
	}

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		self, _, err := cc.SelfApi.GetUser(context.Background())
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("organization_id", self.Id)
	}

	application, _, err := cc.OrganisationApi.AddApplicationByOrga(context.Background(), organizationID.(string), wannabeApplication)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(application.Id)

	return diags
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceApplicationRead(ctx, d, m)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
