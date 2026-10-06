package nodegroup

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.dev/sdk/models"
)

func TestNodegroupFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &KubernetesNodegroup{}

	state.FromNodeGroup(t.Context(), &models.NodeGroup{
		Name: "tf-test-ng", Flavor: "M", TargetNodeCount: 3, ClusterID: "cluster_1",
	}, &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	if state.Name.IsNull() || state.Flavor.IsNull() {
		t.Error("name and flavor must be set")
	}
	if state.Size.ValueInt64() != 3 {
		t.Errorf("size = %d, want 3", state.Size.ValueInt64())
	}
}

// id lives in the Terraform identity, which is this resource's source of truth,
// so the mapper must not touch it.
func TestNodegroupFromAPI_LeavesTheIdentityAlone(t *testing.T) {
	var diags diag.Diagnostics
	state := &KubernetesNodegroup{
		ID:           types.StringValue("from-identity"),
		KubernetesID: types.StringValue("from-state"),
	}

	state.FromNodeGroup(t.Context(), &models.NodeGroup{
		Name: "n", Flavor: "M", TargetNodeCount: 1, ClusterID: "something-else",
	}, &diags)

	if state.ID.ValueString() != "from-identity" {
		t.Errorf("id = %q, want the identity value", state.ID.ValueString())
	}
	if state.KubernetesID.ValueString() != "from-state" {
		t.Errorf("kubernetes_id = %q, want the prior state value", state.KubernetesID.ValueString())
	}
}

func TestNodegroupFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &KubernetesNodegroup{Name: types.StringValue("kept-name")}

	state.FromNodeGroup(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Name.ValueString() != "kept-name" {
		t.Error("name must be preserved")
	}
}
