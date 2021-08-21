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
				Type:     schema.TypeString,
				Default:  "par",
				Optional: true,
			},
			"deploy_type": {
				Type:     schema.TypeString,
				Default:  "git",
				Optional: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"min_instances": {
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
			},
			"max_instances": {
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
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
			"separate_build": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"build_flavor": {
				Type:         schema.TypeString,
				Optional:     true,
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
		Zone:            d.Get("zone").(string),
		Deploy:          d.Get("deploy_type").(string),
		InstanceType:    instance.Type,
		InstanceVersion: instance.Version,
		InstanceVariant: instance.Variant.Id,
		MinInstances:    int32(d.Get("min_instances").(int)),
		MaxInstances:    int32(d.Get("max_instances").(int)),
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

	if homogeneous, ok := d.GetOk("homogeneous"); ok {
		wannabeApplication.Homogeneous = homogeneous.(bool)
	}

	if stickySessions, ok := d.GetOk("sticky_sessions"); ok {
		wannabeApplication.StickySessions = stickySessions.(bool)
	}

	if cancelOnPush, ok := d.GetOk("cancel_on_push"); ok {
		wannabeApplication.CancelOnPush = cancelOnPush.(bool)
	}

	if forceHTTPS, ok := d.GetOk("force_https"); ok {
		if forceHTTPS.(bool) {
			wannabeApplication.ForceHttps = "ENABLED"
		} else {
			wannabeApplication.ForceHttps = "DISABLED"
		}
	}

	if favorite, ok := d.GetOk("favorite"); ok {
		wannabeApplication.Favourite = favorite.(bool)
	}

	if separateBuild, ok := d.GetOk("separate_build"); ok {
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

	var application clevercloud.ApplicationView

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		self, _, err := cc.SelfApi.GetUser(context.Background())
		if err != nil {
			return diag.Errorf("[Create]SelfApi.GetUser: %s\n", err.Error())
		}

		d.Set("organization_id", self.Id)
		application, _, err = cc.SelfApi.AddSelfApplication(context.Background(), wannabeApplication)
		if err != nil {
			return diag.Errorf("[Create]SelfApi.AddSelfApplication: %s\n%s\n", err.Error(), string(err.(clevercloud.GenericOpenAPIError).Body()))
		}
	} else {
		var err error
		application, _, err = cc.OrganisationApi.AddApplicationByOrga(context.Background(), organizationID.(string), wannabeApplication)
		if err != nil {
			return diag.Errorf("[Create]OrganisationApi.AddApplicationByOrga: %s\n", err.Error())
		}
	}

	d.SetId(application.Id)

	resourceApplicationRead(ctx, d, m)

	return diags
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	var application clevercloud.ApplicationView
	var tags []string

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		self, _, err := cc.SelfApi.GetUser(context.Background())
		if err != nil {
			return diag.Errorf("[Read]SelfApi.GetUser: %s\n", err.Error())
		}

		d.Set("organization_id", self.Id)

		if application, _, err = cc.SelfApi.GetSelfApplicationByAppId(context.Background(), d.Id()); err != nil {
			return diag.Errorf("[Create]SelfApi.GetSelfApplicationByAppId: %s\n", err.Error())
		}

		tags, _, err = cc.SelfApi.GetSelfApplicationTagsByAppId(context.Background(), d.Id())
		if err != nil {
			return diag.Errorf("[Create]SelfApi.GetSelfApplicationTagsByAppId: %s\n", err.Error())
		}
	} else {
		var err error
		if application, _, err = cc.OrganisationApi.GetApplicationByOrgaAndAppId(context.Background(), organizationID.(string), d.Id()); err != nil {
			return diag.Errorf("[Create]OrganisationApi.GetApplicationByOrgaAndAppId: %s\n", err.Error())
		}

		tags, _, err = cc.OrganisationApi.GetApplicationTagsByOrgaAndAppId(context.Background(), d.Get("organization_id").(string), d.Id())
		if err != nil {
			return diag.Errorf("[Create]OrganisationApi.GetApplicationTagsByOrgaAndAppId: %s\n", err.Error())
		}
	}

	d.Set("name", application.Name)
	d.Set("description", application.Description)
	d.Set("type", application.Instance.Type)
	d.Set("zone", application.Zone)
	d.Set("deploy_type", application.Deployment.Type)
	d.Set("min_instances", application.Instance.MinInstances)
	d.Set("max_instances", application.Instance.MaxInstances)
	d.Set("min_flavor", application.Instance.MinFlavor.Name)
	d.Set("max_flavor", application.Instance.MaxFlavor.Name)
	d.Set("tags", tags)
	d.Set("archived", application.Archived)
	d.Set("homogeneous", application.Homogeneous)
	d.Set("sticky_sessions", application.StickySessions)
	d.Set("cancel_on_push", application.CancelOnPush)

	d.Set("force_https", application.ForceHttps)
	if application.ForceHttps == "ENABLED" {
		d.Set("force_https", true)
	} else {
		d.Set("force_https", false)
	}

	d.Set("favorite", application.Favourite)
	d.Set("separate_build", application.SeparateBuild)
	d.Set("build_flavor", application.BuildFlavor.Name)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	description, ok := d.GetOk("description")
	if !ok {
		description = d.Get("name")
	}

	instance, errors := getInstanceByType(cc, d.Get("type").(string))
	if errors != nil {
		return errors
	}

	defaultFlavorName := instance.DefaultFlavor.Name

	wannabeApplication := clevercloud.WannabeApplication{
		Name:            d.Get("name").(string),
		Description:     description.(string),
		Zone:            d.Get("zone").(string),
		Deploy:          d.Get("deploy_type").(string),
		InstanceType:    instance.Type,
		InstanceVersion: instance.Version,
		InstanceVariant: instance.Variant.Id,
		MinInstances:    int32(d.Get("min_instances").(int)),
		MaxInstances:    int32(d.Get("max_instances").(int)),
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
		stringTags := make([]string, 0)
		for _, tag := range tags.([]interface{}) {
			stringTags = append(stringTags, tag.(string))
		}
		wannabeApplication.Tags = stringTags
	}

	if homogeneous, ok := d.GetOk("homogeneous"); ok {
		wannabeApplication.Homogeneous = homogeneous.(bool)
	}

	if stickySessions, ok := d.GetOk("sticky_sessions"); ok {
		wannabeApplication.StickySessions = stickySessions.(bool)
	}

	if cancelOnPush, ok := d.GetOk("cancel_on_push"); ok {
		wannabeApplication.CancelOnPush = cancelOnPush.(bool)
	}

	if forceHTTPS, ok := d.GetOk("force_https"); ok {
		if forceHTTPS.(bool) {
			wannabeApplication.ForceHttps = "ENABLED"
		} else {
			wannabeApplication.ForceHttps = "DISABLED"
		}
	}

	if favorite, ok := d.GetOk("favorite"); ok {
		wannabeApplication.Favourite = favorite.(bool)
	}

	if separateBuild, ok := d.GetOk("separate_build"); ok {
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

		if _, _, err = cc.SelfApi.EditSelfApplicationByAppId(context.Background(), d.Id(), wannabeApplication); err != nil {
			return diag.FromErr(err)
		}
	} else {
		var err error
		if _, _, err = cc.OrganisationApi.EditApplicationByOrgaAndAppId(context.Background(), organizationID.(string), d.Id(), wannabeApplication); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceApplicationRead(ctx, d, m)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
