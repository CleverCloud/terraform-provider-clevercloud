package provider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
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

//go:embed resource_nodejs.md
var nodejsDoc string

func (r ResourceNodeJS) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {

	res.Schema = schema.Schema{
		MarkdownDescription: nodejsDoc,
		Attributes: attributes.WithRuntimeCommons(map[string]schema.Attribute{
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
		}),
	}
}
