package clevercloud

import (
	"context"
	"sort"

	"github.com/clevercloud/clevercloud-go/clevercloud"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type resourceApplicationType struct{}

func (r resourceApplicationType) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"type": {
				Type:     types.StringType,
				Required: true,
			},
			"zone": {
				Type:     types.StringType,
				Optional: true,
			},
			"deploy_type": {
				Type:     types.StringType,
				Optional: true,
			},
			"organization_id": {
				Type:     types.StringType,
				Optional: true,
			},
			// "scalability": {
			// 	Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
			// 		"min_instances": {
			// 			Type:     types.NumberType,
			// 			Optional: true,
			// 		},
			// 		"max_instances": {
			// 			Type:     types.NumberType,
			// 			Optional: true,
			// 		},
			// 		"min_flavor": {
			// 			Type:     types.StringType,
			// 			Optional: true,
			// 		},
			// 		"max_flavor": {
			// 			Type:     types.StringType,
			// 			Optional: true,
			// 		},
			// 		"max_allowed_instances": {
			// 			Type:     types.NumberType,
			// 			Computed: true,
			// 		},
			// 	}),
			// 	Optional: true,
			// },
			// "properties": {
			// 	Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
			// 		"homogeneous": {
			// 			Type:     types.BoolType,
			// 			Optional: true,
			// 		},
			// 		"sticky_sessions": {
			// 			Type:     types.BoolType,
			// 			Optional: true,
			// 		},
			// 		"cancel_on_push": {
			// 			Type:     types.BoolType,
			// 			Optional: true,
			// 		},
			// 		"force_https": {
			// 			Type:     types.BoolType,
			// 			Optional: true,
			// 		},
			// 	}),
			// 	Optional: true,
			// },
			"build": {
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					"separate_build": {
						Type:     types.BoolType,
						Optional: true,
					},
					"build_flavor": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
				Optional: true,
			},
			// "environment": {
			// 	Type: types.MapType{
			// 		ElemType: types.StringType,
			// 	},
			// 	Optional: true,
			// },
			// "exposed_environment": {
			// 	Type: types.MapType{
			// 		ElemType: types.StringType,
			// 	},
			// 	Optional: true,
			// },
			// "dependencies": {
			// 	Type: types.MapType{
			// 		ElemType: types.StringType,
			// 	},
			// 	Optional: true,
			// },
			// "vhosts": {
			// 	Type: types.ListType{
			// 		ElemType: types.StringType,
			// 	},
			// 	Optional: true,
			// },
			"favorite": {
				Type:     types.BoolType,
				Optional: true,
			},
			"archived": {
				Type:     types.BoolType,
				Optional: true,
			},
			"tags": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
		},
	}, nil
}

func (r resourceApplicationType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, []*tfprotov6.Diagnostic) {
	return resourceApplication{
		p: *(p.(*provider)),
	}, nil
}

type resourceApplication struct {
	p provider
}

func (app Application) attributeDescriptionOrDefault() string {
	if app.Description.Null {
		return app.Name.Value
	}

	return app.Description.Value
}

func (app Application) attributeZoneOrDefault() string {
	if app.Zone.Null {
		return "par"
	}

	return app.Description.Value
}

func (app Application) attributeDeployTypeOrDefault() string {
	if app.DeployType.Null {
		return "GIT"
	}

	return app.DeployType.Value
}

func getInstanceByType(cc *clevercloud.APIClient, instanceType string) (*clevercloud.AvailableInstanceView, error) {
	instances, _, err := cc.ProductsApi.GetAvailableInstances(context.Background(), &clevercloud.GetAvailableInstancesOpts{})
	if err != nil {
		return nil, err
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

// func (app Application) attributeForceHttpsToString() string {
// 	if app.Properties.ForceHTTPS.Value {
// 		return "ENABLED"
// 	} else {
// 		return "DISABLED"
// 	}
// }

func forceHttpsToBool(value string) bool {
	if value == "ENABLED" {
		return true
	} else {
		return false
	}
}

func (r resourceApplication) applicationModelToWannabeApplication(ctx context.Context, plan *Application) (*clevercloud.WannabeApplication, *tfprotov6.Diagnostic) {
	planApplicationInstance, err := getInstanceByType(r.p.client, plan.Type.Value)
	if err != nil {
		return nil, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error fetching instance type from plan",
			Detail:   "An unexpected error was encountered while reading the plan: " + err.Error(),
		}
	}

	defaultFlavorName := planApplicationInstance.DefaultFlavor.Name
	// planMinInstance, _ := plan.Scalability.MinInstances.Value.Int64()
	// planMaxInstance, _ := plan.Scalability.MaxInstances.Value.Int64()

	wannabeApplication := clevercloud.WannabeApplication{
		Name:            plan.Name.Value,
		Description:     plan.attributeDescriptionOrDefault(),
		Zone:            plan.attributeZoneOrDefault(),
		Deploy:          plan.attributeDeployTypeOrDefault(),
		InstanceType:    plan.Type.Value,
		InstanceVersion: planApplicationInstance.Version,
		InstanceVariant: planApplicationInstance.Variant.Id,
		// MinInstances:    int32(planMinInstance),
		// MaxInstances:    int32(planMaxInstance),
		// MinFlavor:       defaultFlavorName,
		// MaxFlavor:       defaultFlavorName,
		// Homogeneous:     plan.Properties.Homogeneous.Value,
		// StickySessions:  plan.Properties.StickySessions.Value,
		// CancelOnPush:    plan.Properties.CancelOnPush.Value,
		// ForceHttps:      plan.attributeForceHttpsToString(),
		SeparateBuild: plan.Build.SeparateBuild.Value,
		BuildFlavor:   defaultFlavorName,
		Favourite:     plan.Favorite.Value,
		Archived:      plan.Archived.Value,
	}

	for _, flavor := range planApplicationInstance.Flavors {
		// if plan.Scalability.MinFlavor.Value == flavor.Name {
		// 	wannabeApplication.MinFlavor = flavor.Name
		// }
		// if plan.Scalability.MaxFlavor.Value == flavor.Name {
		// 	wannabeApplication.MaxFlavor = flavor.Name
		// }
		if plan.Build.SeparateBuild.Value && plan.Build.BuildFlavor.Value == flavor.Name {
			wannabeApplication.BuildFlavor = flavor.Name
		}
	}

	if err := plan.Tags.ElementsAs(ctx, wannabeApplication.Tags, false); err != nil {
		return nil, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error interfacing tags from plan",
			Detail:   "An unexpected error was encountered while reading the plan: " + err.Error(),
		}
	}

	return &wannabeApplication, nil
}

func applicationViewToApplicationModel(application *clevercloud.ApplicationView) *Application {
	return &Application{
		ID:             types.String{Value: application.Id},
		Name:           types.String{Value: application.Name},
		Description:    types.String{Value: application.Description},
		Type:           types.String{Value: application.Instance.Type},
		Zone:           types.String{Value: application.Zone},
		DeployType:     types.String{Value: application.Deployment.Type},
		OrganizationID: types.String{Value: application.OwnerId},
		// Scalability: ApplicationScalability{
		// 	MinInstances:        types.Number{Value: big.NewFloat(float64(application.Instance.MinInstances))},
		// 	MaxInstances:        types.Number{Value: big.NewFloat(float64(application.Instance.MaxInstances))},
		// 	MinFlavor:           types.String{Value: application.Instance.MinFlavor.Name},
		// 	MaxFlavor:           types.String{Value: application.Instance.MaxFlavor.Name},
		// 	MaxAllowedInstances: types.Number{Value: big.NewFloat(float64(application.Instance.MaxAllowedInstances))},
		// },
		// Properties: ApplicationProperties{
		// 	Homogeneous:    types.Bool{Value: application.Homogeneous},
		// 	StickySessions: types.Bool{Value: application.StickySessions},
		// 	CancelOnPush:   types.Bool{Value: application.CancelOnPush},
		// 	ForceHTTPS:     types.Bool{Value: forceHttpsToBool(application.ForceHttps)},
		// },
		Build: ApplicationBuild{
			SeparateBuild: types.Bool{Value: application.SeparateBuild},
			BuildFlavor:   types.String{Value: application.BuildFlavor.Name},
		},
		Favorite: types.Bool{Value: application.Favourite},
		Archived: types.Bool{Value: application.Archived},
		Tags:     types.List{ElemType: types.StringType},
	}
}

func (r resourceApplication) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Provider not configured",
			Detail:   "The provider hasn't been configured before apply.",
		})
		return
	}

	var plan Application
	if err := req.Plan.Get(ctx, &plan); err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error reading plan",
			Detail:   "An unexpected error was encountered while reading the plan: " + err.Error(),
		})
		return
	}

	wannabeApplication, err := r.applicationModelToWannabeApplication(ctx, &plan)
	if err != nil {
		resp.Diagnostics = append(resp.Diagnostics, err)
		return
	}

	var application clevercloud.ApplicationView
	var tags []string

	if !plan.OrganizationID.Null {
		_, _, err := r.p.client.SelfApi.GetUser(context.Background())
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while fetching self user",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}

		application, _, err = r.p.client.SelfApi.AddSelfApplication(context.Background(), *wannabeApplication)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while creating application for self",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}

		tags, _, err = r.p.client.SelfApi.GetSelfApplicationTagsByAppId(context.Background(), application.Id)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while fetching application tags for self",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}
	} else {
		var err error
		application, _, err = r.p.client.OrganisationApi.AddApplicationByOrga(context.Background(), plan.OrganizationID.Value, *wannabeApplication)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while creating application for organization: " + plan.OrganizationID.Value,
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}

		tags, _, err = r.p.client.OrganisationApi.GetApplicationTagsByOrgaAndAppId(context.Background(), plan.OrganizationID.Value, application.Id)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while fetching application tags for organization",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}
	}

	var result = applicationViewToApplicationModel(&application)

	for _, tag := range tags {
		result.Tags.Elems = append(result.Tags.Elems, types.String{Value: tag})
	}

	if err := resp.State.Set(ctx, result); err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error setting application state",
			Detail:   "Could not set state, unexpected error: " + err.Error(),
		})
		return
	}
}

func (r resourceApplication) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state Application
	if err := req.State.Get(ctx, &state); err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error reading state",
			Detail:   "An unexpected error was encountered while reading the state: " + err.Error(),
		})
		return
	}

	applicationID := state.ID.Value

	var application clevercloud.ApplicationView
	var tags []string

	if !state.OrganizationID.Null {
		_, _, err := r.p.client.SelfApi.GetUser(context.Background())
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while fetching self user",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}

		application, _, err = r.p.client.SelfApi.GetSelfApplicationByAppId(context.Background(), applicationID)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while reading application for self",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}

		tags, _, err = r.p.client.SelfApi.GetSelfApplicationTagsByAppId(context.Background(), application.Id)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while fetching application tags for self",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}
	} else {
		var err error
		application, _, err = r.p.client.OrganisationApi.GetApplicationByOrgaAndAppId(context.Background(), state.OrganizationID.Value, applicationID)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while reading application for organization: " + state.OrganizationID.Value,
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}

		tags, _, err = r.p.client.OrganisationApi.GetApplicationTagsByOrgaAndAppId(context.Background(), state.OrganizationID.Value, application.Id)
		if err != nil {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
				Severity: tfprotov6.DiagnosticSeverityError,
				Summary:  "Request error while fetching application tags for organization",
				Detail:   "An unexpected error was encountered while requesting the api: " + err.Error(),
			})
			return
		}
	}

	var result = applicationViewToApplicationModel(&application)

	for _, tag := range tags {
		result.Tags.Elems = append(result.Tags.Elems, types.String{Value: tag})
	}

	if err := resp.State.Set(ctx, result); err != nil {
		resp.Diagnostics = append(resp.Diagnostics, &tfprotov6.Diagnostic{
			Severity: tfprotov6.DiagnosticSeverityError,
			Summary:  "Error setting state",
			Detail:   "Unexpected error encountered trying to set new state: " + err.Error(),
		})
		return
	}
}

func (r resourceApplication) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

func (r resourceApplication) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}
