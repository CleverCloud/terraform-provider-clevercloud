package nodegroup

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/sdk/models"
)

type KubernetesNodegroup struct {
	ID           types.String `tfsdk:"id"`
	KubernetesID types.String `tfsdk:"kubernetes_id"`
	Name         types.String `tfsdk:"name"`
	Flavor       types.String `tfsdk:"flavor"`
	Size         types.Int64  `tfsdk:"size"`
}

type KubernetesNodegroupIdentity struct {
	ID types.String `tfsdk:"id"`
}

//go:embed doc.md
var resourceKubernetesNodegroupDoc string

func (r ResourceKubernetesNodegroup) IdentitySchema(_ context.Context, req resource.IdentitySchemaRequest, res *resource.IdentitySchemaResponse) {
	res.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r ResourceKubernetesNodegroup) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKubernetesNodegroupDoc,
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"kubernetes_id": schema.StringAttribute{Required: true, MarkdownDescription: "ID of the Kubernetes cluster"},
			"name":          schema.StringAttribute{Required: true, MarkdownDescription: "Name of the node group"},
			"flavor":        schema.StringAttribute{Required: true, MarkdownDescription: "Size of the nodes, see [available flavors](https://www.clever.cloud/developers/doc/kubernetes/)"},
			"size":          schema.Int64Attribute{Required: true, MarkdownDescription: "Number of nodes"}, // add validator with min=0, max=16
		},
	}
}

// The API-to-state mappers for clevercloud_kubernetes_nodegroup follow.
// See CONTRIBUTING.md § "API → state mapping".

// FromNodeGroup maps the node-group view.
//
// id is deliberately not mapped: this resource keeps it in the Terraform
// identity, which is its source of truth, and the CRUD copies it from there.
func (ng *KubernetesNodegroup) FromNodeGroup(ctx context.Context, api *models.NodeGroup, diags *diag.Diagnostics) {
	if ng == nil || api == nil {
		return
	}

	ng.Name = pkg.FromStr(api.Name)
	ng.Flavor = pkg.FromStr(string(api.Flavor))
	ng.Size = pkg.FromI(api.TargetNodeCount)
}
