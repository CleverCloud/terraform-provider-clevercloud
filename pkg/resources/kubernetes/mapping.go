package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_kubernetes.
// See CONTRIBUTING.md § "API → state mapping".

// FromCluster maps the cluster view, which the create, the read and the cluster
// poll all answer with — tmp.KubernetesInfo and tmp.KubernetesCreateResponse
// are both aliases of tmp.ClusterView, so one mapper serves all three.
//
// id is deliberately not mapped here: this resource keeps it in the Terraform
// identity, which is its source of truth, and the CRUD copies it from there.
func (k *Kubernetes) FromCluster(ctx context.Context, api *tmp.ClusterView, diags *diag.Diagnostics) *Kubernetes {
	if k == nil || api == nil {
		return k
	}

	k.Name = pkg.FromStr(api.Name)
	k.NodeAutoprovisioning = pkg.FromBool(autoprovisioningFeatureEnabled(api.Features))

	return k
}

// FromKubeconfig maps the separate kubeconfig call.
func (k *Kubernetes) FromKubeconfig(ctx context.Context, kubeconfig *string, diags *diag.Diagnostics) *Kubernetes {
	if k == nil || kubeconfig == nil {
		return k
	}

	k.KubeConfig = pkg.FromStr(*kubeconfig)

	return k
}
