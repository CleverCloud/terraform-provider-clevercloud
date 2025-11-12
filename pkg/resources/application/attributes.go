package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type VHost struct {
	FQDN      types.String `tfsdk:"fqdn"`
	PathBegin types.String `tfsdk:"path_begin"`
}

func (vh VHost) String() *string {
	if vh.FQDN.IsNull() || vh.FQDN.IsUnknown() {
		return nil
	}

	path := "/"
	if !vh.PathBegin.IsNull() && !vh.PathBegin.IsUnknown() {
		path = vh.PathBegin.ValueString()
	}

	vhost := fmt.Sprintf("%s%s", vh.FQDN.ValueString(), path)
	return &vhost
}

type Runtime struct {
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	Description      types.String           `tfsdk:"description"`
	MinInstanceCount types.Int64            `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64            `tfsdk:"max_instance_count"`
	SmallestFlavor   types.String           `tfsdk:"smallest_flavor"`
	BiggestFlavor    types.String           `tfsdk:"biggest_flavor"`
	BuildFlavor      types.String           `tfsdk:"build_flavor"`
	Region           types.String           `tfsdk:"region"`
	StickySessions   types.Bool             `tfsdk:"sticky_sessions"`
	RedirectHTTPS    types.Bool             `tfsdk:"redirect_https"`
	VHosts           types.Set              `tfsdk:"vhosts"`
	DeployURL        types.String           `tfsdk:"deploy_url"`
	Dependencies     types.Set              `tfsdk:"dependencies"`
	Networkgroups    types.Set              `tfsdk:"networkgroups"`
	Deployment       *attributes.Deployment `tfsdk:"deployment"`
	Hooks            *attributes.Hooks      `tfsdk:"hooks"`

	// Env
	AppFolder   types.String `tfsdk:"app_folder"`
	Environment types.Map    `tfsdk:"environment"`
}

type RuntimeV0 struct {
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	Description      types.String           `tfsdk:"description"`
	MinInstanceCount types.Int64            `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64            `tfsdk:"max_instance_count"`
	SmallestFlavor   types.String           `tfsdk:"smallest_flavor"`
	BiggestFlavor    types.String           `tfsdk:"biggest_flavor"`
	BuildFlavor      types.String           `tfsdk:"build_flavor"`
	Region           types.String           `tfsdk:"region"`
	StickySessions   types.Bool             `tfsdk:"sticky_sessions"`
	RedirectHTTPS    types.Bool             `tfsdk:"redirect_https"`
	VHosts           types.Set              `tfsdk:"vhosts"`
	DeployURL        types.String           `tfsdk:"deploy_url"`
	Dependencies     types.Set              `tfsdk:"dependencies"`
	Deployment       *attributes.Deployment `tfsdk:"deployment"`
	Hooks            *attributes.Hooks      `tfsdk:"hooks"`

	// Env
	AppFolder   types.String `tfsdk:"app_folder"`
	Environment types.Map    `tfsdk:"environment"`
}

func (r Runtime) DependenciesAsString(ctx context.Context, diags *diag.Diagnostics) []string {
	dependencies := []string{}
	diags.Append(r.Dependencies.ElementsAs(ctx, &dependencies, false)...)
	return dependencies
}

func (r Runtime) VHostsAsStrings(ctx context.Context, diags *diag.Diagnostics) []string {
	vhosts := pkg.SetTo[VHost](ctx, r.VHosts, diags)
	if diags.HasError() {
		return []string{}
	}

	return pkg.Reduce(vhosts, []string{}, func(acc []string, s VHost) []string {
		str := s.String()
		if str == nil {
			return acc
		}

		acc = append(acc, *str)
		return acc
	})
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
		MarkdownDescription: "Maximum instance count, if different from min value, enable auto-scaling",
	},
	"smallest_flavor": schema.StringAttribute{
		Required:            true,
		Validators:          []validator.String{helper.UpperCaseValidator},
		MarkdownDescription: "Smallest instance flavor",
	},
	"biggest_flavor": schema.StringAttribute{
		Required:            true,
		Validators:          []validator.String{helper.UpperCaseValidator},
		MarkdownDescription: "Biggest instance flavor, if different from smallest, enable auto-scaling",
	},
	"build_flavor": schema.StringAttribute{
		Optional:            true,
		Validators:          []validator.String{helper.UpperCaseValidator},
		MarkdownDescription: "Use dedicated instance with given flavor for build phase",
	},
	"region": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("par"),
		MarkdownDescription: "Geographical region where the database will be deployed",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	},
	"sticky_sessions": schema.BoolAttribute{
		Optional:            true,
		MarkdownDescription: "Enable sticky sessions, use it when your client sessions are instances scoped",
		Default:             booldefault.StaticBool(false),
		Computed:            true,
	},
	"redirect_https": schema.BoolAttribute{
		Optional:            true,
		MarkdownDescription: "Redirect client from plain to TLS port",
		Default:             booldefault.StaticBool(false),
		Computed:            true,
	},
	"vhosts": schema.SetNestedAttribute{
		MarkdownDescription: "List of virtual hosts",
		Optional:            true,
		Computed:            true, // needed to set cleverapps defautlt domains
		PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"fqdn": schema.StringAttribute{
					Required:            true,
					PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					MarkdownDescription: "Fully qualified domain name",
					Validators: []validator.String{
						pkg.NewValidatorRegex("Validate domain format", pkg.VhostValidRegExp),
					},
				},
				"path_begin": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString("/"),
					PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					MarkdownDescription: "Any HTTP request starting with this path will be sent to this application",
					Validators: []validator.String{
						pkg.NewValidator("Path must start with /", func(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
							if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
								return
							}
							value := req.ConfigValue.ValueString()
							if !strings.HasPrefix(value, "/") {
								res.Diagnostics.AddAttributeError(
									req.Path,
									"Invalid path_begin format",
									fmt.Sprintf("path_begin must start with '/' (got: '%s')", value),
								)
							}
						}),
					},
				},
			},
		},
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
		MarkdownDescription: "A list of application or add-ons required to run this application.\nCan be either app_xxx or postgres_yyy ID format",
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
						res.Diagnostics.AddError("This dependency doesn't have a valid format", fmt.Sprintf("'%s' is neither an App ID or a Real ID", item))
					}
				}
			}),
		},
	},
	"networkgroups": schema.SetNestedAttribute{
		Optional:            true,
		MarkdownDescription: "List of networkgroups the application must be part of",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"networkgroup_id": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "ID of the networkgroup",
				},
				"fqdn": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "domain name which will resolve to application instances inside the networkgroup",
					//Default: stringdefault.StaticString(),
				},
			},
		},
	},
}

var runtimeCommonV0 = map[string]schema.Attribute{
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
		MarkdownDescription: "Maximum instance count, if different from min value, enable auto-scaling",
	},
	"smallest_flavor": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Smallest instance flavor",
	},
	"biggest_flavor": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Biggest instance flavor, if different from smallest, enable auto-scaling",
	},
	"build_flavor": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Use dedicated instance with given flavor for build phase",
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
	"vhosts": schema.SetAttribute{
		MarkdownDescription: "List of virtual hosts",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true, // needed to set cleverapps defautlt domains
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
		MarkdownDescription: "A list of application or add-ons required to run this application.\nCan be either app_xxx or postgres_yyy ID format",
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
						res.Diagnostics.AddError("This dependency doesn't have a valid format", fmt.Sprintf("'%s' is neither an App ID or a Real ID", item))
					}
				}
			}),
		},
	},
}

func WithRuntimeCommons(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(runtimeCommon, runtimeSpecifics)
}

func WithRuntimeCommonsV0(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(runtimeCommonV0, runtimeSpecifics)
}
