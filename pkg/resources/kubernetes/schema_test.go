package kubernetes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// The kubeconfig gives administrator access to the cluster: Terraform must
// redact it from plans and outputs.
func TestKubeconfigIsSensitive(t *testing.T) {
	resp := &resource.SchemaResponse{}
	ResourceKubernetes{}.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["kubeconfig"]
	if !ok {
		t.Fatal("kubeconfig is missing from the schema")
	}
	if !attr.IsSensitive() {
		t.Error("kubeconfig is not sensitive")
	}
}
