package elasticsearch_cluster

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func apiCluster(nodes ...tmp.ElasticsearchNode) *tmp.ElasticsearchCluster {
	return &tmp.ElasticsearchCluster{
		ID: "cluster_1", Name: "tf-test-cluster", Username: "user",
		NetworkGroupID: "ng_1",
		Version:        tmp.ElasticsearchVersion{Major: 8, Minor: 11, Patch: 3},
		Nodes:          nodes,
	}
}

func TestElasticsearchClusterFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &ElasticsearchCluster{}

	state.FromCluster(t.Context(), apiCluster(
		tmp.ElasticsearchNode{Plan: &tmp.ElasticsearchPlan{Name: "xs"}},
	), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	for name, value := range map[string]types.String{
		"id": state.ID, "name": state.Name, "username": state.Username,
		"network_group_id": state.NetworkGroupID, "plan": state.Plan,
	} {
		if value.IsNull() || value.IsUnknown() {
			t.Errorf("%s is %v, want a value", name, value)
		}
	}
	if state.NodeCount.ValueInt64() != 1 {
		t.Errorf("node_count = %d, want the number of nodes", state.NodeCount.ValueInt64())
	}
	if state.Version.IsNull() {
		t.Error("version must be set")
	}
}

// The plan is only echoed back per node, so a cluster whose nodes are not listed
// yet must keep the configured value rather than lose it.
func TestElasticsearchClusterFromAPI_PreservesWhatTheAPIDoesNotReturn(t *testing.T) {
	var diags diag.Diagnostics
	state := &ElasticsearchCluster{
		Plan:     types.StringValue("xs"),
		Password: types.StringValue("kept-password"),
	}

	state.FromCluster(t.Context(), apiCluster(), &diags)

	if state.Plan.ValueString() != "xs" {
		t.Error("plan must be preserved while the nodes are not listed")
	}
	if state.Password.ValueString() != "kept-password" {
		t.Error("the cluster view carries no password and must not clear one")
	}
}

// The credentials endpoint answers empty while the cluster boots, and never
// answers the password again afterwards.
func TestElasticsearchClusterFromAPI_EmptyCredentialsKeepState(t *testing.T) {
	var diags diag.Diagnostics
	state := &ElasticsearchCluster{
		Username: types.StringValue("kept-user"),
		Password: types.StringValue("kept-password"),
	}

	state.FromCredentials(t.Context(), &tmp.ElasticsearchCredentials{}, &diags)

	if state.Username.ValueString() != "kept-user" || state.Password.ValueString() != "kept-password" {
		t.Error("empty credentials must leave state alone")
	}
}

func TestElasticsearchClusterFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &ElasticsearchCluster{Name: types.StringValue("kept-name")}

	state.FromCluster(t.Context(), nil, &diags)
	state.FromCredentials(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Name.ValueString() != "kept-name" {
		t.Error("name must be preserved")
	}
}
