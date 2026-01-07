package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/attributes"
)

// VHost represents a virtual host configuration
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

// Runtime represents the common fields for all application runtimes
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
	Networkgroups    types.Set                 `tfsdk:"networkgroups"`
	Deployment       *attributes.Deployment    `tfsdk:"deployment"`
	Hooks            *attributes.Hooks         `tfsdk:"hooks"`
	Integrations     *attributes.Integrations  `tfsdk:"integrations"`

	// Env
	AppFolder          types.String `tfsdk:"app_folder"`
	Environment        types.Map    `tfsdk:"environment"`
	ExposedEnvironment types.Map    `tfsdk:"exposed_environment"`
}

// RuntimeV0 represents the schema v0 of Runtime (for state upgrades)
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
	// If vhosts is null or unknown, return nil to indicate "not specified"
	// This allows SyncVHostsOnCreate to keep the default vhosts from the API
	if r.VHosts.IsNull() || r.VHosts.IsUnknown() {
		return nil
	}

	vhosts := pkg.SetTo[VHost](ctx, r.VHosts, diags)
	if diags.HasError() {
		return []string{}
	}

	// If no vhosts are present, return an empty slice (not nil)
	// This distinguishes "explicitly empty" from "not specified"
	items := []string{}
	for _, vhost := range vhosts {
		s := vhost.String()
		if s != nil {
			items = append(items, *s)
		}
	}

	return items
}

// GetRuntimePtr returns a pointer to the Runtime struct for modification
func (r *Runtime) GetRuntimePtr() *Runtime {
	return r
}
