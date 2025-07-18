package attributes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
)

type Runtime struct {
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
	VHosts           types.Set    `tfsdk:"vhosts"`
	AdditionalVHosts types.List   `tfsdk:"additional_vhosts"`
	DeployURL        types.String `tfsdk:"deploy_url"`
	Dependencies     types.Set    `tfsdk:"dependencies"`
	Deployment       *Deployment  `tfsdk:"deployment"`
	Hooks            *Hooks       `tfsdk:"hooks"`

	// Env
	AppFolder   types.String `tfsdk:"app_folder"`
	Environment types.Map    `tfsdk:"environment"`
}

func (r Runtime) VHostsAsStrings(ctx context.Context, diags *diag.Diagnostics) []string {
	vhosts := []string{}
	diags.Append(r.VHosts.ElementsAs(ctx, &vhosts, true)...)
	return vhosts
}

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
		//Default:             int64default.StaticInt64(1), // TODO: setup all defaults
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
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("par"),
		MarkdownDescription: "Geographical region where the database will be deployed",
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
		DeprecationMessage:  "Use vhosts instead (if you provide any vhost, no cleverapps.io domain will be returned)",
	},
	"vhosts": schema.SetAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Add custom hostname, see [documentation](https://www.clever-cloud.com/doc/administrate/domain-names/)",
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
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"deploy_url": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Git URL used to push source code",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"environment": schema.MapAttribute{
		Optional:    true,
		Sensitive:   true,
		Description: "Environment variables injected into the application",
		ElementType: types.StringType,
	},

	"dependencies": schema.SetAttribute{
		Optional:            true,
		MarkdownDescription: "A list of application or addons requires to run this application.\nCan be either app_xxx or postgres_yyy ID format",
		ElementType:         types.StringType,
		Validators: []validator.Set{
			pkg.NewSetValidator("Check ID format", func(ctx context.Context, req validator.SetRequest, res *validator.SetResponse) {
				if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
					return
				}

				items := []types.String{}
				res.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &items, false)...)
				if res.Diagnostics.HasError() {
					return
				}

				knownItems := pkg.Filter(items, func(item types.String) bool { return !item.IsUnknown() })
				stringItems := pkg.Map(knownItems, func(item types.String) string { return item.ValueString() })

				for _, item := range stringItems {
					if !pkg.AddonRegExp.MatchString(item) &&
						!pkg.AppRegExp.MatchString(item) &&
						!pkg.ServiceRegExp.MatchString(item) {
						res.Diagnostics.AddError("This dependecy don't have a valid format", fmt.Sprintf("'%s' is neither an App ID or an addon ID", item))
					}
				}
			}),
		},
	},
}

func WithRuntimeCommons(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(runtimeCommon, runtimeSpecifics)
}
