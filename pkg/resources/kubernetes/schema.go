package kubernetes

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// nodeAutoprovisioningDefault is what the schema gives the attribute when the
// configuration leaves it out, and what Create and Update fall back to when
// they have to materialise that default themselves
const nodeAutoprovisioningDefault = false

type Kubernetes struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	KubeConfig           types.String `tfsdk:"kubeconfig"`
	NodeAutoprovisioning types.Bool   `tfsdk:"node_autoprovisioning"`
}

type KubernetesIdentity struct {
	ID types.String `tfsdk:"id"`
}

//go:embed doc.md
var resourceKubernetesDoc string

func (r ResourceKubernetes) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             0,
		MarkdownDescription: resourceKubernetesDoc,
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Generated unique identifier", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":       schema.StringAttribute{Required: true, MarkdownDescription: "Name of the Kubernetes cluster"},
			"kubeconfig": schema.StringAttribute{Computed: true, MarkdownDescription: "Kubernetes configuration file content for accessing the cluster"},
			"node_autoprovisioning": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(nodeAutoprovisioningDefault),
				MarkdownDescription: "Enable node autoscaling via node auto-provisioning, powered by Karpenter. " +
					"Nodes are created and deleted on demand from the `NodePool` and `CleverNodeClass` custom resources you deploy in the cluster",
			},
		},
	}
}

func (r ResourceKubernetes) IdentitySchema(_ context.Context, req resource.IdentitySchemaRequest, res *resource.IdentitySchemaResponse) {
	res.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

// The API-to-state mappers for clevercloud_kubernetes follow.
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
