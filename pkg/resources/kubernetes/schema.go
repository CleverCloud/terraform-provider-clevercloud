package kubernetes

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
				Default:  booldefault.StaticBool(false),
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
