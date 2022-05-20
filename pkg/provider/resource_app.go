package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// App is a generic App config, common for all runtime
// It should not be use as is, but be extended to a specific Runtime
// Not usable for now, see https://github.com/hashicorp/terraform-plugin-framework/issues/309
type App struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	MinInstanceCount types.Int64  `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64  `tfsdk:"max_instance_count"`
	SmallestFlavor   types.String `tfsdk:"smallest_flavor"`
	BiggestFlavor    types.String `tfsdk:"biggest_flavor"`
	BuildFlavor      types.String `tfsdk:"build_flavor"`
	Region           types.String `tfsdk:"region"`
	StickySessions   types.Bool   `tfsdk:"sticky_sessions"`
	RedirectHTTPS    types.Bool   `tfsdk:"redirect_https"`
	Commit           types.String `tfsdk:"commit"`
	VHost            types.String `tfsdk:"vhost"`
	AdditionalVHosts types.List   `tfsdk:"additional_vhosts"`
	DeployURL        types.String `tfsdk:"deploy_url"`
	AppFolder        types.String `tfsdk:"app_folder"`
	Environment      types.Map    `tfsdk:"environment"`
	Dependencies     types.List   `tfsdk:"dependencies"`
}

func GetAppSchemaAttributes() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		// Plan provided

		"name": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Application name",
		},
		"description": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Application description",
		},
		"min_instance_count": {
			Type:                types.Int64Type,
			Required:            true,
			MarkdownDescription: "Minimum instance count",
		},
		"max_instance_count": {
			Type:                types.Int64Type,
			Required:            true,
			MarkdownDescription: "Maximum instance count, if different from min value, enable autoscaling",
		},
		"smallest_flavor": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Smallest instance flavor",
		},
		"biggest_flavor": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Biggest intance flavor, if different from smallest, enable autoscaling",
		},
		"build_flavor": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Use dedicated instance with given flavor for build step",
		},
		"region": {
			Type:                types.StringType,
			Required:            true,
			MarkdownDescription: "Geographical region where the app will be deployed",
		},
		"sticky_sessions": {
			Type:                types.BoolType,
			Optional:            true,
			MarkdownDescription: "Enable sticky sessions, use it when your client sessions are instances scoped",
		},
		"redirect_https": {
			Type:                types.BoolType,
			Optional:            true,
			MarkdownDescription: "Redirect client from plain to TLS port",
		},
		"commit": {
			Type:                types.StringType,
			Optional:            true,
			Description:         "Support either '<branch>:<SHA>' or '<tag>'",
			MarkdownDescription: "Deploy application on the given commit/tag",
		},
		"additional_vhosts": {
			Type:                types.ListType{ElemType: types.StringType},
			Optional:            true,
			MarkdownDescription: "Add custom hostname in addition to the default one, see [documentation](https://www.clever-cloud.com/doc/administrate/domain-names/)",
		},
		"environment": {
			Type:                types.MapType{ElemType: types.StringType},
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Environment variables to set on the application",
		},
		"dependencies": {
			Type:                types.ListType{ElemType: types.StringType},
			Optional:            true,
			MarkdownDescription: "List of service IDs required to run this app",
		},
		// APP_FOLDER
		"app_folder": {
			Type:                types.StringType,
			Optional:            true,
			MarkdownDescription: "Folder in which the application is located (inside the git repository)",
		},

		// API provided
		// provider provided

		"id": {
			Type:                types.StringType,
			Computed:            true,
			MarkdownDescription: "Unique identifier generated during application creation",
		},
		"deploy_url": {
			Type:                types.StringType,
			Computed:            true,
			MarkdownDescription: "Git URL used to push source code",
		},
		// cleverapps one
		"vhost": {
			Type:                types.StringType,
			Computed:            true,
			MarkdownDescription: "Default vhost to access your app",
		},
	}
}

// Use the plan to compute environment variables to set on the application
// use custom env vars, then add application specific ones
// We should support all section described here
// https://www.clever-cloud.com/doc/reference/reference-environment-variables/#variables-you-can-define
func (app App) GetEnv(ctx context.Context) (map[string]string, diag.Diagnostics) {
	m := map[string]string{}
	diags := diag.Diagnostics{}

	// plan defined environment variables
	for k, item := range app.Environment.Elems {
		var strValue string
		diags.Append(tfsdk.ValueAs(ctx, item, &strValue)...)
		m[k] = strValue
	}

	if app.AppFolder.Value != "" {
		m["APP_FOLDER"] = app.AppFolder.Value
	}

	return m, diags
}
