package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodeJS struct {
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
	DevDependencies  types.Bool   `tfsdk:"dev_dependencies"`
	StartScript      types.String `tfsdk:"start_script"`
	PackageManager   types.String `tfsdk:"package_manager"`
	Registry         types.String `tfsdk:"registry"`
	RegistryToken    types.String `tfsdk:"registry_token"`
}

const nodejsDoc = `
Manage [NodeJS](https://nodejs.org/) applications.

See [NodeJS product](https://www.clever-cloud.com/nodejs-hosting/) specification.
`

func (r ResourceNodeJS) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		MarkdownDescription: nodejsDoc,
		Attributes: map[string]schema.Attribute{
			// customer provided
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Application name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Application description",
			},
			"min_instance_count": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Minimum instance count",
			},
			"max_instance_count": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Maximum instance count, if different from min value, enable autoscaling",
			},
			"smallest_flavor": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Smallest instance flavor",
			},
			"biggest_flavor": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Biggest intance flavor, if different from smallest, enable autoscaling",
			},
			"build_flavor": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Use dedicated instance with given flavor for build step",
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Geographical region where the app will be deployed",
			},
			"sticky_sessions": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable sticky sessions, use it when your client sessions are instances scoped",
			},
			"redirect_https": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Redirect client from plain to TLS port",
			},
			"commit": schema.StringAttribute{
				Optional:            true,
				Description:         "Support either '<branch>:<SHA>' or '<tag>'",
				MarkdownDescription: "Deploy application on the given commit/tag",
			},
			"additional_vhosts": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Add custom hostname in addition to the default one, see [documentation](https://www.clever-cloud.com/doc/administrate/domain-names/)",
			},
			// APP_FOLDER
			"app_folder": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Folder in which the application is located (inside the git repository)",
			},

			// Node specifique

			// CC_NODE_DEV_DEPENDENCIES
			"dev_dependencies": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Install development dependencies specified in package.json",
			},
			// CC_RUN_COMMAND
			"start_script": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Set custom start script, instead of `npm start`",
			},
			// CC_NODE_BUILD_TOOL / CC_CUSTOM_BUILD_TOOL
			"package_manager": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Either npm, npm-ci, yarn, yarn2 or custom",
			},
			// CC_NPM_REGISTRY
			"registry": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The host of your private repository, available values: github or the registry host",
			},
			// NPM_TOKEN
			"registry_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Private repository token",
			},

			// provider provided

			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier generated during application creation",
			},
			"deploy_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Git URL used to push source code",
			},
			// cleverapps one
			"vhost": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Default vhost to access your app",
			},
		},
	}
}
