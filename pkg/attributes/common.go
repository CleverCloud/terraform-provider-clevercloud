package attributes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This attributes are used on several runtimes
var runtimeCommon = map[string]schema.Attribute{
	// client provided

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
	"additional_vhosts": schema.ListAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		MarkdownDescription: "Add custom hostname in addition to the default one, see [documentation](https://www.clever-cloud.com/doc/administrate/domain-names/)",
	},
	"commit": schema.StringAttribute{
		Optional:            true,
		Description:         "Support either '<branch>:<SHA>' or '<tag>'",
		MarkdownDescription: "Deploy application on the given commit/tag",
	},
	// APP_FOLDER
	"app_folder": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Folder in which the application is located (inside the git repository)",
	},

	// Provider provided
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

	"environment": schema.MapAttribute{
		Optional:    true,
		Sensitive:   true,
		Description: "Environment variables injected into the application",
		ElementType: types.StringType,
	},
}

func WithRuntimeCommons(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	m := map[string]schema.Attribute{}

	for attrName, attr := range runtimeCommon {
		m[attrName] = attr
	}

	for attrName, attr := range runtimeSpecifics {
		m[attrName] = attr
	}

	return m
}
