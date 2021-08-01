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

	d.Set("name", application.Name)
	d.Set("description", application.Description)
	d.Set("zone", application.Zone)

	defaultEnv := make(map[string]interface{}, len(application.Instance.DefaultEnv))
	for key, value := range application.Instance.DefaultEnv {
		defaultEnv[key] = value
	}

	instanceBindings := &schema.Set{F: schema.HashResource(applicationInstanceResource)}
	instanceBindings.Add(map[string]interface{}{
		"type":                  application.Instance.Type,
		"version":               application.Instance.Version,
		"variant":               makeInstanceVariantResourceSchemaSet(&application.Instance.Variant),
		"min_instances":         int(application.Instance.MinInstances),
		"max_instances":         int(application.Instance.MaxInstances),
		"max_allowed_instances": int(application.Instance.MaxAllowedInstances),
		"min_flavor":            makeFlavorResourceSchemaSet(&application.Instance.MinFlavor),
		"max_flavor":            makeFlavorResourceSchemaSet(&application.Instance.MaxFlavor),
		"flavors":               makeFlavorsResourceSchemaList(application.Instance.Flavors),
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

	d.Set("archived", application.Archived)
	d.Set("sticky_sessions", application.StickySessions)
	d.Set("homogeneous", application.Homogeneous)
	d.Set("favorite", application.Favourite)
	d.Set("cancel_on_push", application.CancelOnPush)
	d.Set("webhook_url", application.WebhookUrl)
	d.Set("webhook_secret", application.WebhookSecret)
	d.Set("separate_build", application.SeparateBuild)

	if err := d.Set("build_flavor", makeFlavorResourceSchemaSet(&application.BuildFlavor)); err != nil {
		return diag.FromErr(fmt.Errorf("cannot set application build flavor bindings (%s): %v", d.Id(), err))
	}

	d.Set("state", application.State)
	d.Set("commit_id", application.CommitId)
	d.Set("appliance", application.Appliance)
	d.Set("branch", application.Branch)
	d.Set("force_https", application.ForceHttps)
	d.Set("deploy_url", application.DeployUrl)
	d.Set("owner_id", application.OwnerId)

	return diags
}

func makeInstanceVariantResourceSchemaSet(instanceVariant *clevercloud.InstanceVariantView) *schema.Set {
	set := &schema.Set{F: schema.HashResource(applicationInstanceVariantResource)}

	set.Add(map[string]interface{}{
		"id":          instanceVariant.Id,
		"slug":        instanceVariant.Slug,
		"name":        instanceVariant.Name,
		"deploy_type": instanceVariant.DeployType,
		"logo":        instanceVariant.Logo,
	})

	return set
}

func makeFlavorsResourceSchemaList(flavors []clevercloud.FlavorView) []interface{} {
	list := make([]interface{}, 0)

	for _, flavor := range flavors {
		list = append(list, makeFlavorResourceSchemaSet(&flavor))
	}

	return list
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
