package kubernetes

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.dev/sdk/models"
)

// apiTopology is the control plane topology of a cluster as the API spells it.
// ALL_IN_ONE and DEDICATED_COMPUTE take a flavor and a replication factor,
// DISTRIBUTED takes both for each component instead.
//
// The SDK publishes it as models.TopologyConfig, a union holding raw JSON and
// no typed member: this is the typed view of it.
type apiTopology struct {
	Topology          string                 `json:"topology"`
	Flavor            string                 `json:"flavor,omitempty"`
	ReplicationFactor int64                  `json:"replicationFactor,omitempty"`
	Components        *apiTopologyComponents `json:"components,omitempty"`
}

// apiTopologyComponents holds the settings of each control plane component of a
// DISTRIBUTED topology
type apiTopologyComponents struct {
	Apiserver              apiTopologyComponent `json:"apiserver"`
	CloudControllerManager apiTopologyComponent `json:"cloudControllerManager"`
	ControllerManager      apiTopologyComponent `json:"controllerManager"`
	NodeGroupOperator      apiTopologyComponent `json:"nodeGroupOperator"`
	Scheduler              apiTopologyComponent `json:"scheduler"`
}

// apiTopologyComponent is the size of one control plane component
type apiTopologyComponent struct {
	Flavor            string `json:"flavor"`
	ReplicationFactor int64  `json:"replicationFactor"`
}

// toSDK wraps the topology in the union the SDK sends, nil when there is none
func (t *apiTopology) toSDK(diags *diag.Diagnostics) *models.TopologyConfig {
	if t == nil {
		return nil
	}

	raw, err := json.Marshal(t)
	if err != nil {
		diags.AddError("failed to encode the topology of the kubernetes cluster", err.Error())
		return nil
	}

	config := &models.TopologyConfig{}
	if err := config.UnmarshalJSON(raw); err != nil {
		diags.AddError("failed to encode the topology of the kubernetes cluster", err.Error())
		return nil
	}

	return config
}

// apiTopologyFrom reads the union the SDK received, nil when the API gave no
// topology
func apiTopologyFrom(config models.TopologyConfig, diags *diag.Diagnostics) *apiTopology {
	raw, err := config.MarshalJSON()
	if err != nil {
		diags.AddError("failed to read the topology of the kubernetes cluster", err.Error())
		return nil
	}

	var topology *apiTopology
	if err := json.Unmarshal(raw, &topology); err != nil {
		diags.AddError("failed to read the topology of the kubernetes cluster", err.Error())
		return nil
	}

	return topology
}
