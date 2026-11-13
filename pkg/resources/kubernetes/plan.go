package kubernetes

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
)

var _ resource.ResourceWithModifyPlan = (*ResourceKubernetes)(nil)

// ModifyPlan aligns the topology with its configuration and checks the requested
// region, version and topology against what the platform offers, so a refused
// value fails the plan rather than the apply
func (r *ResourceKubernetes) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	plan := helper.From[Kubernetes](ctx, req.Plan, &resp.Diagnostics)
	config := helper.From[Kubernetes](ctx, req.Config, &resp.Diagnostics)
	var state *Kubernetes
	if !req.State.Raw.IsNull() {
		state = helper.From[Kubernetes](ctx, req.State, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// the nested attributes being computed, the plan holds unknowns or leftovers
	// of the previous topology for the ones the configuration leaves out. They
	// have no value in that case: the fields of a topology are exclusive
	if isSet(config.Topology) {
		plan.Topology = config.Topology
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("topology"), config.Topology)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	replaced := state != nil && isSet(config.Topology) && topologyChanged(plan.Topology, state.Topology)
	if replaced {
		resp.RequiresReplace.Append(path.Root("topology"))
		resp.Diagnostics.AddAttributeWarning(
			path.Root("topology"),
			"changing the topology replaces the kubernetes cluster",
			"the API cannot change the topology of a cluster: this plan deletes it, with every workload and volume it holds, and creates a new one",
		)
	}

	checkVersion, checkTopology := planChecks(plan, config, state)

	if checkVersion && state != nil && !replaced && versionLower(plan.Version.ValueString(), state.Version.ValueString()) {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("version"),
			"this version change is a downgrade",
			fmt.Sprintf("moving the cluster from %s to %s is a downgrade, which Kubernetes does not support: objects stored by the newer version may not be readable by the older one",
				state.Version.ValueString(), plan.Version.ValueString()),
		)
	}

	if r.Provider == nil {
		return
	}

	if regionCheck(plan, config, state) {
		r.checkRegion(ctx, plan.Region.ValueString(), &resp.Diagnostics)
	}

	if !checkVersion && !checkTopology {
		return
	}

	product := r.SDK.V4().KubernetesProduct().Getkubernetesserviceconfig(ctx)
	if product.HasError() {
		// not a reason to block the plan, the API still validates on apply
		tflog.Warn(ctx, "failed to get the kubernetes service configuration, skipping validation", map[string]any{"error": product.Error().Error()})
		return
	}

	if checkVersion {
		validateVersion(plan.Version.ValueString(), product.Payload().Versions, &resp.Diagnostics)
	}
	if checkTopology {
		validateTopology(ctx, plan.Topology, product.Payload().Topologies, &resp.Diagnostics)
	}
}

// planChecks tells which attributes have to be validated. A value that does not
// change must not be: a cluster keeps working on a version the platform stopped
// offering, and that must not fail every later plan. Only a configured value is
// checked, a computed one comes from the API or from the state
func planChecks(plan, config, state *Kubernetes) (checkVersion, checkTopology bool) {
	checkVersion = isSet(config.Version) && isSet(plan.Version) &&
		(state == nil || !plan.Version.Equal(state.Version))
	checkTopology = isSet(config.Topology) && isSet(plan.Topology) &&
		(state == nil || !plan.Topology.Equal(state.Topology))

	return checkVersion, checkTopology
}

// regionCheck tells whether the region has to be validated, under the same rule
// as planChecks: only a configured value that creates or moves the cluster. A
// cluster keeps running in a zone the organisation may no longer create in, and
// setting a region on a state without one is regionOnUnknownZone's to refuse
func regionCheck(plan, config, state *Kubernetes) bool {
	return isSet(config.Region) && isSet(plan.Region) &&
		(state == nil || (isSet(state.Region) && !plan.Region.Equal(state.Region)))
}

// checkRegion validates the region against the zones the organisation may create
// a cluster in, so a refused one fails the plan rather than the apply
func (r *ResourceKubernetes) checkRegion(ctx context.Context, region string, diags *diag.Diagnostics) {
	ctx, cancel := context.WithTimeout(ctx, zoneLookupTimeout)
	defer cancel()

	zones := tmp.ListKubernetesZones(ctx, r.Client(), r.Organization())
	if zones.HasError() || zones.Payload() == nil {
		// not a reason to block the plan, the API still validates on apply
		tflog.Warn(ctx, "failed to list the kubernetes zones, skipping the validation of the region", map[string]any{"error": fmt.Sprint(zones.Error())})
		return
	}

	validateRegion(region, *zones.Payload(), diags)
}

// validateRegion refuses a region that names none of the zones, or several of
// them: zone names are not unique, and the API refuses a name it cannot resolve
// to a single zone
func validateRegion(region string, zones []tmp.KubernetesZone, diags *diag.Diagnostics) {
	matches := 0
	names := make([]string, 0, len(zones))
	for _, zone := range zones {
		if zone.Name == region {
			matches++
		}
		if !slices.Contains(names, zone.Name) {
			names = append(names, zone.Name)
		}
	}
	slices.Sort(names)

	switch {
	case matches == 1:
	case matches > 1:
		diags.AddAttributeError(
			path.Root("region"),
			"ambiguous kubernetes region",
			fmt.Sprintf("%d of the zones this organisation may create a cluster in are named '%s', and the API refuses a name that designates several zones", matches, region),
		)
	case len(names) == 0:
		diags.AddAttributeError(
			path.Root("region"),
			"invalid kubernetes region",
			fmt.Sprintf("region '%s' is not available, this organisation may not create a cluster in any region", region),
		)
	default:
		diags.AddAttributeError(
			path.Root("region"),
			"invalid kubernetes region",
			fmt.Sprintf("region '%s' is not available, available regions are: %s", region, strings.Join(names, ", ")),
		)
	}
}

// topologyChanged reports whether the plan replaces the cluster. A state holding
// null says nothing about the real topology, and deleting a cluster takes more
// than not knowing: the next refresh tells, and the plan after it can decide
func topologyChanged(plan, state types.Object) bool {
	return isSet(plan) && isSet(state) && !plan.Equal(state)
}

// versionLower reports whether a is an older MAJOR.MINOR than b. A version it
// cannot parse is not reported as one
func versionLower(a, b string) bool {
	aMajor, aMinor, aOk := parseVersion(a)
	bMajor, bMinor, bOk := parseVersion(b)
	if !aOk || !bOk {
		return false
	}

	return aMajor < bMajor || (aMajor == bMajor && aMinor < bMinor)
}

func parseVersion(version string) (major, minor int, ok bool) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}

	return major, minor, true
}

func validateVersion(version string, versions models.VersionsConfig, diags *diag.Diagnostics) {
	if slices.Contains(versions.Available, version) {
		return
	}

	diags.AddAttributeError(
		path.Root("version"),
		"invalid kubernetes version",
		fmt.Sprintf("version '%s' is not available, available versions are: %s", version, strings.Join(versions.Available, ", ")),
	)
}

// validateTopology checks flavors and replication factors against the
// constraints of the requested topology type. The shape of the attribute is
// validateTopologyShape's concern
func validateTopology(ctx context.Context, value types.Object, topologies []models.TopologyConstraints, diags *diag.Diagnostics) {
	var topology Topology
	diags.Append(value.As(ctx, &topology, asOptions)...)
	if diags.HasError() || !isSet(topology.Type) {
		return
	}

	index := slices.IndexFunc(topologies, func(t models.TopologyConstraints) bool {
		return t.Topology.String() == topology.Type.ValueString()
	})
	if index < 0 {
		return
	}
	constraints := topologies[index]
	root := path.Root("topology")

	if !isSet(topology.Components) {
		validateSize(root, TopologyComponent{Flavor: topology.Flavor, ReplicationFactor: topology.ReplicationFactor}, constraints, diags)
		return
	}

	var components TopologyComponents
	diags.Append(topology.Components.As(ctx, &components, asOptions)...)
	if diags.HasError() {
		return
	}

	for _, component := range components.all() {
		validateSize(root.AtName("components").AtName(component.name), component.decode(ctx, diags), constraints, diags)
	}
}

func validateSize(at path.Path, size TopologyComponent, constraints models.TopologyConstraints, diags *diag.Diagnostics) {
	if isSet(size.Flavor) {
		allowed := pkg.Map(constraints.AvailableFlavors, func(f models.NodeFlavor) string { return f.String() })
		if !slices.Contains(allowed, size.Flavor.ValueString()) {
			diags.AddAttributeError(
				at.AtName("flavor"),
				"invalid topology flavor",
				fmt.Sprintf("flavor '%s' is not allowed for %s, allowed flavors are: %s", size.Flavor.ValueString(), constraints.Topology, strings.Join(allowed, ", ")),
			)
		}
	}

	if isSet(size.ReplicationFactor) {
		factor, bounds := size.ReplicationFactor.ValueInt64(), constraints.ReplicationFactor
		if factor < int64(bounds.Min) || factor > int64(bounds.Max) {
			diags.AddAttributeError(
				at.AtName("replication_factor"),
				"invalid topology replication factor",
				fmt.Sprintf("replication factor %d is out of bounds, it has to be between %d and %d", factor, bounds.Min, bounds.Max),
			)
		}
	}
}
