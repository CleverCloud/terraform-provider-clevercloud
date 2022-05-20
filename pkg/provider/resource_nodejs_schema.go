package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type NodeJS struct {
	// Runtime common properties
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

	// NodeJS properties
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
	StartScript     types.String `tfsdk:"start_script"`
	PackageManager  types.String `tfsdk:"package_manager"`
	Registry        types.String `tfsdk:"registry"`
	RegistryToken   types.String `tfsdk:"registry_token"`
}

const nodejsDoc = `
Manage [NodeJS](https://nodejs.org/) applications.

See [NodeJS product](https://www.clever-cloud.com/nodejs-hosting/) specification.
`

func (r resourceNodejsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: nodejsDoc,
		Attributes: pkg.MergeMap(GetAppSchemaAttributes(), map[string]tfsdk.Attribute{
			// customer provided
			// Node specifique

			// CC_NODE_DEV_DEPENDENCIES
			"dev_dependencies": {
				Type:                types.BoolType,
				Optional:            true,
				MarkdownDescription: "Install development dependencies specified in package.json",
			},
			// CC_RUN_COMMAND
			"start_script": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Set custom start script, instead of `npm start`",
			},
			// CC_NODE_BUILD_TOOL / CC_CUSTOM_BUILD_TOOL
			"package_manager": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Either npm, npm-ci, yarn, yarn2 or custom",
			},
			// CC_NPM_REGISTRY
			"registry": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "The host of your private repository, available values: github or the registry host",
			},
			// NPM_TOKEN
			"registry_token": {
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Private repository token",
			},
		}),
	}, nil
}

func (plan NodeJS) App() App {
	return App{
		ID:               plan.ID,
		Name:             plan.Name,
		Description:      plan.Description,
		MinInstanceCount: plan.MinInstanceCount,
		MaxInstanceCount: plan.MaxInstanceCount,
		SmallestFlavor:   plan.SmallestFlavor,
		BiggestFlavor:    plan.BiggestFlavor,
		Region:           plan.Region,
		StickySessions:   plan.StickySessions,
		RedirectHTTPS:    plan.RedirectHTTPS,
		Commit:           plan.Commit,
		VHost:            plan.VHost,
		AdditionalVHosts: plan.AdditionalVHosts,
		DeployURL:        plan.DeployURL,
		AppFolder:        plan.AppFolder,
		Environment:      plan.Environment,
	}
}

// Use the plan to compute environment variables to set on the application
// use underlying generic app envs then add app type specific ones
// We should support all section described here
// https://www.clever-cloud.com/doc/reference/reference-environment-variables/#nodejs
func (plan NodeJS) GetEnv(ctx context.Context) (map[string]string, diag.Diagnostics) {
	m, diags := plan.App().GetEnv(ctx)

	if plan.DevDependencies.Value {
		m["CC_NODE_DEV_DEPENDENCIES"] = "install"
	}

	if plan.PackageManager.Value != "" {
		if pkg.NewSet("npm", "npm-ci", "yarn", "yarn2").Contains(plan.PackageManager.Value) {
			m["CC_NODE_BUILD_TOOL"] = plan.PackageManager.Value
		} else {
			m["CC_CUSTOM_BUILD_TOOL"] = plan.PackageManager.Value
		}
	}

	if plan.Registry.Value != "" {
		m["CC_NPM_REGISTRY"] = plan.Region.Value
	}

	if plan.RegistryToken.Value != "" {
		m["NPM_TOKEN"] = plan.RegistryToken.Value
	}

	return m, diags
}
