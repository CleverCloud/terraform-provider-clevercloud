package clevercloud

import (
	"context"
	"fmt"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var applicationInstanceVariantResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"slug": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"deploy_type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"logo": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var applicationInstanceResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
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
		"min_instances": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"max_instances": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"max_allowed_instances": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"min_flavor": {
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     applicationFlavorResource,
		},
		"max_flavor": {
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     applicationFlavorResource,
		},
		"flavors": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeSet,
				Elem: applicationFlavorResource,
			},
		},
		"default_env": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"lifetime": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"instance_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var applicationDeploymentResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"shutdownable": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"repo_state": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"url": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"http_url": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var applicationVhostResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"fqdn": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var applicationFlavorMemoryResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"unit": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"value": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"formatted": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var applicationFlavorResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"mem": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"memory": {
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     applicationFlavorMemoryResource,
		},
		"cpus": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"gpus": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"disk": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"price": {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		"available": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"microservice": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"machine_learning": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"nice": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"price_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"rbd_image": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

func dataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     applicationInstanceResource,
			},
			"deployment": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     applicationDeploymentResource,
			},
			"vhosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeSet,
					Elem: applicationVhostResource,
				},
			},
			"archived": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"sticky_sessions": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"homogeneous": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"favorite": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cancel_on_push": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"force_https": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"separate_build": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"build_flavor": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     applicationFlavorResource,
			},
			"webhook_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"webhook_secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"commit_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"appliance": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deploy_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cc := m.(*clevercloud.APIClient)

	var diags diag.Diagnostics

	var application clevercloud.ApplicationView

	organizationID, ok := d.GetOk("organization_id")
	if !ok {
		self, _, err := cc.SelfApi.GetUser(context.Background())
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("organization_id", self.Id)

		if application, _, err = cc.SelfApi.GetSelfApplicationByAppId(context.Background(), d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		var err error
		if application, _, err = cc.OrganisationApi.GetApplicationByOrgaAndAppId(context.Background(), organizationID.(string), d.Get("id").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(application.Id)

	_ = d.Set("name", application.Name)
	_ = d.Set("description", application.Description)
	_ = d.Set("zone", application.Zone)

	instanceVariantBindings := &schema.Set{F: schema.HashResource(applicationInstanceVariantResource)}
	instanceVariantBindings.Add(map[string]interface{}{
		"id":          application.Instance.Variant.Id,
		"slug":        application.Instance.Variant.Slug,
		"name":        application.Instance.Variant.Name,
		"deploy_type": application.Instance.Variant.DeployType,
		"logo":        application.Instance.Variant.Logo,
	})

	flavors := make([]interface{}, len(application.Instance.Flavors))
	for i, flavor := range application.Instance.Flavors {
		flavors[i] = makeFlavorResourceSchemaSet(&flavor)
	}

	defaultEnv := make(map[string]interface{}, len(application.Instance.DefaultEnv))
	for key, value := range application.Instance.DefaultEnv {
		defaultEnv[key] = value
	}

	instanceBindings := &schema.Set{F: schema.HashResource(applicationInstanceResource)}
	instanceBindings.Add(map[string]interface{}{
		"type":                  application.Instance.Type,
		"version":               application.Instance.Version,
		"variant":               instanceVariantBindings,
		"min_instances":         int(application.Instance.MinInstances),
		"max_instances":         int(application.Instance.MaxInstances),
		"max_allowed_instances": int(application.Instance.MaxAllowedInstances),
		"min_flavor":            makeFlavorResourceSchemaSet(&application.Instance.MinFlavor),
		"max_flavor":            makeFlavorResourceSchemaSet(&application.Instance.MaxFlavor),
		"flavors":               flavors,
		"default_env":           defaultEnv,
		"lifetime":              application.Instance.Lifetime,
		"instance_version":      application.Instance.InstanceAndVersion,
	})

	if err := d.Set("instance", instanceBindings); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set application instance bindings (%s): %v", d.Id(), err))
	}

	deploymentBindings := &schema.Set{F: schema.HashResource(applicationDeploymentResource)}
	deploymentBindings.Add(map[string]interface{}{
		"shutdownable": application.Deployment.Shutdownable,
		"type":         application.Deployment.Type,
		"repo_state":   application.Deployment.RepoState,
		"url":          application.Deployment.Url,
		"http_url":     application.Deployment.HttpUrl,
	})

	if err := d.Set("deployment", deploymentBindings); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set application deployment bindings (%s): %v", d.Id(), err))
	}

	vhostsBindings := make([]interface{}, len(application.Vhosts))
	for i, vhost := range application.Vhosts {
		v := &schema.Set{F: schema.HashResource(applicationVhostResource)}
		v.Add(map[string]interface{}{
			"fqdn": vhost.Fqdn,
		})
		vhostsBindings[i] = v
	}

	if err := d.Set("vhosts", vhostsBindings); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set application vhosts bindings (%s): %v", d.Id(), err))
	}

	_ = d.Set("archived", application.Archived)
	_ = d.Set("sticky_sessions", application.StickySessions)
	_ = d.Set("homogeneous", application.Homogeneous)
	_ = d.Set("favorite", application.Favourite)
	_ = d.Set("cancel_on_push", application.CancelOnPush)
	_ = d.Set("webhook_url", application.WebhookUrl)
	_ = d.Set("webhook_secret", application.WebhookSecret)
	_ = d.Set("separate_build", application.SeparateBuild)

	if err := d.Set("build_flavor", makeFlavorResourceSchemaSet(&application.BuildFlavor)); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set application build flavor bindings (%s): %v", d.Id(), err))
	}

	_ = d.Set("state", application.State)
	_ = d.Set("commit_id", application.CommitId)
	_ = d.Set("appliance", application.Appliance)
	_ = d.Set("branch", application.Branch)
	_ = d.Set("force_https", application.ForceHttps)
	_ = d.Set("deploy_url", application.DeployUrl)
	_ = d.Set("owner_id", application.OwnerId)

	return diags
}

func makeFlavorResourceSchemaSet(flavor *clevercloud.FlavorView) *schema.Set {
	set := &schema.Set{F: schema.HashResource(applicationFlavorResource)}

	memory := &schema.Set{F: schema.HashResource(applicationFlavorMemoryResource)}
	memory.Add(map[string]interface{}{
		"unit":      flavor.Memory.Unit,
		"value":     float64(flavor.Memory.Value),
		"formatted": flavor.Memory.Formatted,
	})

	set.Add(map[string]interface{}{
		"name":             flavor.Name,
		"mem":              int(flavor.Mem),
		"memory":           memory,
		"cpus":             int(flavor.Cpus),
		"gpus":             int(flavor.Gpus),
		"disk":             int(flavor.Disk),
		"price":            float64(flavor.Price),
		"available":        flavor.Available,
		"microservice":     flavor.Microservice,
		"machine_learning": flavor.MachineLearning,
		"nice":             int(flavor.Nice),
		"price_id":         flavor.PriceId,
		"rbd_image":        flavor.Rbdimage,
	})

	return set
}
