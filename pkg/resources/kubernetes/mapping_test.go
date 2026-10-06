package kubernetes

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func boolPtr(b bool) *bool { return &b }

func TestKubernetesFromAPI_PopulatesEveryAttribute(t *testing.T) {
	var diags diag.Diagnostics
	state := &Kubernetes{}

	state.FromCluster(t.Context(), &tmp.ClusterView{
		ID: "cluster_1", Name: "tf-test-k8s",
		Features: &tmp.KubernetesFeatures{NodeAutoprovisioning: boolPtr(true)},
	}, &diags)
	state.FromKubeconfig(t.Context(), strPtr("apiVersion: v1"), &diags)

	if diags.HasError() {
		t.Fatalf("mapping reported %v", diags.Errors())
	}
	if state.Name.IsNull() || state.KubeConfig.IsNull() {
		t.Error("name and kubeconfig must be set")
	}
	if !state.NodeAutoprovisioning.ValueBool() {
		t.Error("node_autoprovisioning should follow the reported feature")
	}
}

// id lives in the Terraform identity, which is this resource's source of truth,
// so the mapper must not touch it.
func TestKubernetesFromAPI_LeavesTheIdentityAlone(t *testing.T) {
	var diags diag.Diagnostics
	state := &Kubernetes{ID: types.StringValue("from-identity")}

	state.FromCluster(t.Context(), &tmp.ClusterView{ID: "something-else", Name: "n"}, &diags)

	if state.ID.ValueString() != "from-identity" {
		t.Errorf("id = %q, want the identity value", state.ID.ValueString())
	}
}

func TestKubernetesFromAPI_NilPayloadIsANoOp(t *testing.T) {
	var diags diag.Diagnostics
	state := &Kubernetes{Name: types.StringValue("kept-name")}

	state.FromCluster(t.Context(), nil, &diags)
	state.FromKubeconfig(t.Context(), nil, &diags)

	if diags.HasError() {
		t.Fatalf("a nil payload must not report an error, got %v", diags.Errors())
	}
	if state.Name.ValueString() != "kept-name" {
		t.Error("name must be preserved")
	}
}

func strPtr(s string) *string { return &s }
