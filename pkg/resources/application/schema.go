package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
)

// runtimeCommon defines common schema attributes for all application runtimes
var runtimeCommon = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Application name",
	},
	"description": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Application description",
	},
	"min_instance_count": schema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Minimum instance count",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		//Default:             int64default.StaticInt64(1), // TODO: setup all defaults
	},
	"max_instance_count": schema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Maximum instance count, if different from min value, enable auto-scaling",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	},
	"smallest_flavor": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Smallest instance flavor",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"biggest_flavor": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Biggest instance flavor, if different from smallest, enable auto-scaling",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
	"exposed_environment": schema.MapAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Sensitive:   true,
		Description: "Environment variables other linked applications will be able to use",
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
	"integrations": attributes.IntegrationsAttribute,
}

// runtimeCommonV0 defines common schema attributes for schema version 0 (for state upgrades)
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
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Minimum instance count",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		//Default:             int64default.StaticInt64(1), // TODO: setup all defaults
	},
	"max_instance_count": schema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Maximum instance count, if different from min value, enable auto-scaling",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	},
	"smallest_flavor": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Smallest instance flavor",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"biggest_flavor": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Biggest instance flavor, if different from smallest, enable auto-scaling",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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

// WithRuntimeCommons merges runtime-specific schema attributes with common ones
func WithRuntimeCommons(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(runtimeCommon, runtimeSpecifics)
}

// WithRuntimeCommonsV0 merges runtime-specific schema attributes with common V0 ones
func WithRuntimeCommonsV0(runtimeSpecifics map[string]schema.Attribute) map[string]schema.Attribute {
	return pkg.Merge(runtimeCommonV0, runtimeSpecifics)
}

// ValidateRuntimeFlavors validates that build_flavor, smallest_flavor, and biggest_flavor
// are valid for the given application variant (docker, nodejs, etc.)
func ValidateRuntimeFlavors(ctx context.Context, r provider.Provider, variantSlug string, runtime Runtime, diags *diag.Diagnostics) {
	org := r.Organization()
	instance := LookupInstanceByVariantSlug(ctx, r.Client(), &org, variantSlug, diags)
	if instance == nil {
		diags.AddError("failed to lookup instance", fmt.Sprintf("no instance named '%s'", variantSlug))
		return
	}
	flavorSet := instance.Flavors.NamesAsSet()

	availableFlavors := strings.Join(flavorSet.Slice(), ", ")

	// Validate build_flavor
	if !runtime.BuildFlavor.IsNull() && !runtime.BuildFlavor.IsUnknown() {
		if !flavorSet.Contains(runtime.BuildFlavor.ValueString()) {
			diags.AddAttributeError(
				path.Root("build_flavor"),
				"invalid flavor",
				fmt.Sprintf("available flavors are: %s", availableFlavors),
			)
		}
	}

	// Validate smallest_flavor
	if !runtime.SmallestFlavor.IsNull() && !runtime.SmallestFlavor.IsUnknown() {
		if !flavorSet.Contains(runtime.SmallestFlavor.ValueString()) {
			diags.AddAttributeError(
				path.Root("smallest_flavor"),
				"invalid flavor",
				fmt.Sprintf("available flavors are: %s", availableFlavors),
			)
		}
	}

	// Validate biggest_flavor
	if !runtime.BiggestFlavor.IsNull() && !runtime.BiggestFlavor.IsUnknown() {
		if !flavorSet.Contains(runtime.BiggestFlavor.ValueString()) {
			diags.AddAttributeError(
				path.Root("biggest_flavor"),
				"invalid flavor",
				fmt.Sprintf("available flavors are: %s", availableFlavors),
			)
		}
	}
}
