package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

const (
	topologyAllInOne         = "ALL_IN_ONE"
	topologyDedicatedCompute = "DEDICATED_COMPUTE"
	topologyDistributed      = "DISTRIBUTED"
)

// Topology is the value of the topology attribute
type Topology struct {
	Type              types.String `tfsdk:"type"`
	Flavor            types.String `tfsdk:"flavor"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
	Components        types.Object `tfsdk:"components"`
}

// TopologyComponents is the value of the topology.components attribute
type TopologyComponents struct {
	Apiserver              types.Object `tfsdk:"apiserver"`
	CloudControllerManager types.Object `tfsdk:"cloud_controller_manager"`
	ControllerManager      types.Object `tfsdk:"controller_manager"`
	NodeGroupOperator      types.Object `tfsdk:"node_group_operator"`
	Scheduler              types.Object `tfsdk:"scheduler"`
}

// all lists the components by attribute name, in a stable order
func (c TopologyComponents) all() []namedComponent {
	return []namedComponent{
		{"apiserver", c.Apiserver},
		{"cloud_controller_manager", c.CloudControllerManager},
		{"controller_manager", c.ControllerManager},
		{"node_group_operator", c.NodeGroupOperator},
		{"scheduler", c.Scheduler},
	}
}

type namedComponent struct {
	name  string
	value types.Object
}

// decode reads the size of a component, zero when it is null or unknown
func (c namedComponent) decode(ctx context.Context, diags *diag.Diagnostics) TopologyComponent {
	var component TopologyComponent
	diags.Append(c.value.As(ctx, &component, asOptions)...)

	return component
}

// TopologyComponent is the size of one control plane component
type TopologyComponent struct {
	Flavor            types.String `tfsdk:"flavor"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
}

var (
	topologyComponentAttrTypes = map[string]attr.Type{
		"flavor":             types.StringType,
		"replication_factor": types.Int64Type,
	}
	topologyComponentsAttrTypes = map[string]attr.Type{
		"apiserver":                types.ObjectType{AttrTypes: topologyComponentAttrTypes},
		"cloud_controller_manager": types.ObjectType{AttrTypes: topologyComponentAttrTypes},
		"controller_manager":       types.ObjectType{AttrTypes: topologyComponentAttrTypes},
		"node_group_operator":      types.ObjectType{AttrTypes: topologyComponentAttrTypes},
		"scheduler":                types.ObjectType{AttrTypes: topologyComponentAttrTypes},
	}
	topologyAttrTypes = map[string]attr.Type{
		"type":               types.StringType,
		"flavor":             types.StringType,
		"replication_factor": types.Int64Type,
		"components":         types.ObjectType{AttrTypes: topologyComponentsAttrTypes},
	}
)

// asOptions decodes an unknown nested value to its zero value. Everything
// reading a decoded value checks that it is known first
var asOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

// isSet reports whether a value is both present and known
func isSet(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// topologyToAPI maps the topology attribute to the API payload, nil when the
// attribute has no value and the API picks the topology
func topologyToAPI(ctx context.Context, value types.Object, diags *diag.Diagnostics) *tmp.KubernetesTopology {
	if !isSet(value) {
		return nil
	}

	var topology Topology
	diags.Append(value.As(ctx, &topology, asOptions)...)
	if diags.HasError() {
		return nil
	}

	payload := &tmp.KubernetesTopology{
		Topology:          topology.Type.ValueString(),
		Flavor:            topology.Flavor.ValueString(),
		ReplicationFactor: topology.ReplicationFactor.ValueInt64(),
	}
	if !isSet(topology.Components) {
		return payload
	}

	var components TopologyComponents
	diags.Append(topology.Components.As(ctx, &components, asOptions)...)
	if diags.HasError() {
		return nil
	}

	toAPI := func(value types.Object) tmp.KubernetesTopologyComponent {
		component := namedComponent{value: value}.decode(ctx, diags)

		return tmp.KubernetesTopologyComponent{
			Flavor:            component.Flavor.ValueString(),
			ReplicationFactor: component.ReplicationFactor.ValueInt64(),
		}
	}
	payload.Components = &tmp.KubernetesTopologyComponents{
		Apiserver:              toAPI(components.Apiserver),
		CloudControllerManager: toAPI(components.CloudControllerManager),
		ControllerManager:      toAPI(components.ControllerManager),
		NodeGroupOperator:      toAPI(components.NodeGroupOperator),
		Scheduler:              toAPI(components.Scheduler),
	}

	return payload
}

// topologyFromAPI maps the topology returned by the API to the state value. The
// fields of a topology are exclusive: a value for the ones its type does not
// take would differ from the configuration and replace the cluster
func topologyFromAPI(ctx context.Context, payload *tmp.KubernetesTopology, diags *diag.Diagnostics) types.Object {
	if payload == nil {
		return types.ObjectNull(topologyAttrTypes)
	}

	topology := Topology{
		Type:              pkg.FromStr(payload.Topology),
		Flavor:            types.StringNull(),
		ReplicationFactor: types.Int64Null(),
		Components:        types.ObjectNull(topologyComponentsAttrTypes),
	}

	if payload.Components == nil {
		topology.Flavor = pkg.FromStr(payload.Flavor)
		topology.ReplicationFactor = pkg.FromI(payload.ReplicationFactor)
	} else {
		fromAPI := func(c tmp.KubernetesTopologyComponent) types.Object {
			component, d := types.ObjectValueFrom(ctx, topologyComponentAttrTypes, TopologyComponent{
				Flavor:            pkg.FromStr(c.Flavor),
				ReplicationFactor: pkg.FromI(c.ReplicationFactor),
			})
			diags.Append(d...)

			return component
		}
		components, d := types.ObjectValueFrom(ctx, topologyComponentsAttrTypes, TopologyComponents{
			Apiserver:              fromAPI(payload.Components.Apiserver),
			CloudControllerManager: fromAPI(payload.Components.CloudControllerManager),
			ControllerManager:      fromAPI(payload.Components.ControllerManager),
			NodeGroupOperator:      fromAPI(payload.Components.NodeGroupOperator),
			Scheduler:              fromAPI(payload.Components.Scheduler),
		})
		diags.Append(d...)
		topology.Components = components
	}

	value, d := types.ObjectValueFrom(ctx, topologyAttrTypes, topology)
	diags.Append(d...)

	return value
}

var _ resource.ResourceWithValidateConfig = (*ResourceKubernetes)(nil)

// ValidateConfig enforces the shape of the topology, which depends on its type.
// It needs no API call, so it also runs on terraform validate
func (r *ResourceKubernetes) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	config := helper.From[Kubernetes](ctx, req.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	validateTopologyShape(ctx, config.Topology, &resp.Diagnostics)
}

func validateTopologyShape(ctx context.Context, value types.Object, diags *diag.Diagnostics) {
	if !isSet(value) {
		return
	}

	var topology Topology
	diags.Append(value.As(ctx, &topology, asOptions)...)
	if diags.HasError() {
		return
	}

	root := path.Root("topology")
	if topology.Type.IsNull() {
		diags.AddAttributeError(root.AtName("type"), "missing topology type", "a topology requires a type")
		return
	}
	if topology.Type.IsUnknown() {
		return
	}

	if topology.Type.ValueString() == topologyDistributed {
		if topology.Components.IsNull() {
			diags.AddAttributeError(root.AtName("components"), "missing topology components", "a DISTRIBUTED topology sizes each component, set components")
		}
		validateComponentsShape(ctx, root.AtName("components"), topology.Components, diags)
		if !topology.Flavor.IsNull() {
			diags.AddAttributeError(root.AtName("flavor"), "unexpected topology flavor", "a DISTRIBUTED topology sizes each component, set the flavor in components")
		}
		if !topology.ReplicationFactor.IsNull() {
			diags.AddAttributeError(root.AtName("replication_factor"), "unexpected topology replication factor", "a DISTRIBUTED topology sizes each component, set the replication factor in components")
		}

		return
	}

	if topology.Flavor.IsNull() {
		diags.AddAttributeError(root.AtName("flavor"), "missing topology flavor", topology.Type.ValueString()+" requires a flavor")
	}
	if topology.ReplicationFactor.IsNull() {
		diags.AddAttributeError(root.AtName("replication_factor"), "missing topology replication factor", topology.Type.ValueString()+" requires a replication factor")
	}
	if !topology.Components.IsNull() {
		diags.AddAttributeError(root.AtName("components"), "unexpected topology components", "components only applies to a DISTRIBUTED topology")
	}
}

func validateComponentsShape(ctx context.Context, at path.Path, value types.Object, diags *diag.Diagnostics) {
	if !isSet(value) {
		return
	}

	var components TopologyComponents
	diags.Append(value.As(ctx, &components, asOptions)...)
	if diags.HasError() {
		return
	}

	for _, component := range components.all() {
		if component.value.IsNull() {
			diags.AddAttributeError(at.AtName(component.name), "missing topology component", component.name+" requires a flavor and a replication factor")
			continue
		}

		size := component.decode(ctx, diags)
		if isSet(component.value) && size.Flavor.IsNull() {
			diags.AddAttributeError(at.AtName(component.name).AtName("flavor"), "missing topology component flavor", component.name+" requires a flavor")
		}
		if isSet(component.value) && size.ReplicationFactor.IsNull() {
			diags.AddAttributeError(at.AtName(component.name).AtName("replication_factor"), "missing topology component replication factor", component.name+" requires a replication factor")
		}
	}
}
