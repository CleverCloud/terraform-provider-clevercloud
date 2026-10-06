package nodegroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/sdk/models"
)

// mapping.go holds every API-to-state mapper for clevercloud_kubernetes_nodegroup.
// See CONTRIBUTING.md § "API → state mapping".

// FromNodeGroup maps the node-group view.
//
// id is deliberately not mapped: this resource keeps it in the Terraform
// identity, which is its source of truth, and the CRUD copies it from there.
func (ng *KubernetesNodegroup) FromNodeGroup(ctx context.Context, api *models.NodeGroup, diags *diag.Diagnostics) *KubernetesNodegroup {
	if ng == nil || api == nil {
		return ng
	}

	ng.Name = pkg.FromStr(api.Name)
	ng.Flavor = pkg.FromStr(string(api.Flavor))
	ng.Size = pkg.FromI(api.TargetNodeCount)

	return ng
}
