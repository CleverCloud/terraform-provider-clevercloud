package provider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PHP struct {
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

	// PHP related
	PHPVersion      types.String `tfsdk:"php_version"`
	WebRoot         types.String `tfsdk:"webroot"`
	RedisSessions   types.Bool   `tfsdk:"redis_sessions"`
	DevDependencies types.Bool   `tfsdk:"dev_dependencies"`
}

//go:embed resource_php.md
var phpDoc string

func (r ResourcePHP) Schema(ctx context.Context, req resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		MarkdownDescription: phpDoc,
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

			// PHP specifique

			// CC_WEBROOT
			"php_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "PHP version (Default: 7)",
			},
			"webroot": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Define the DocumentRoot of your project (default: \".\")",
			},

			"redis_sessions": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Use a linked Redis instance to store sessions (Default: false)",
			},
			"dev_dependencies": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Install development dependencies",
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

func (php *PHP) toEnv() map[string]string {
	m := map[string]string{}

	IfIsSet(php.WebRoot, func(webroot string) {
		m["CC_WEBROOT"] = webroot
	})
	IfIsSet(php.PHPVersion, func(version string) {
		m["CC_PHP_VERSION"] = version
	})
	IfIsSetB(php.DevDependencies, func(devDeps bool) {
		if devDeps {
			m["CC_PHP_DEV_DEPENDENCIES"] = "install"
		}
	})
	IfIsSetB(php.RedisSessions, func(redis bool) {
		if redis {
			m["SESSION_TYPE"] = "redis"
		}
	})

	return m
}
