package kubernetes

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
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
	PersistentStorage    types.Bool   `tfsdk:"persistent_storage"`
	Tags                 types.Set    `tfsdk:"tags"`
	Topology             types.Object `tfsdk:"topology"`
	Version              types.String `tfsdk:"version"`
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
			"persistent_storage": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Enable the CSI driver that provides persistent volumes. " +
					"The API leaves it disabled on a new cluster unless it is set to `true`. Removing the attribute keeps the current value",
			},
			"tags": resources.TagsAttribute,
			"topology": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				// The nested attributes are computed too, whatever the configuration
				// has to set: Terraform reads a non computed value in the state as
				// proof the attribute was configured and plans its removal when it
				// is absent, which would show a diff on every plan. So their shape
				// is enforced by ValidateConfig, and ModifyPlan aligns the plan with
				// the configuration and decides on the replacement
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Control plane topology, the API defaults to `DEDICATED_COMPUTE` with the `XS` flavor and a replication factor of 1. " +
					"It cannot be changed on an existing cluster: changing it replaces the cluster",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Validators:          []validator.String{stringvalidator.OneOf(topologyAllInOne, topologyDedicatedCompute, topologyDistributed)},
						MarkdownDescription: "Topology type, required: `ALL_IN_ONE` (control plane and workloads share the same VMs), `DEDICATED_COMPUTE` or `DISTRIBUTED`",
					},
					"flavor": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Flavor of the control plane, required unless `type` is `DISTRIBUTED`. The accepted flavors depend on the topology and are checked at plan time",
					},
					"replication_factor": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Replication factor of the control plane, required unless `type` is `DISTRIBUTED`",
					},
					"components": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Flavor and replication factor of each control plane component, required when `type` is `DISTRIBUTED`",
						Attributes: map[string]schema.Attribute{
							"apiserver":                topologyComponentSchema("API server"),
							"cloud_controller_manager": topologyComponentSchema("cloud controller manager"),
							"controller_manager":       topologyComponentSchema("controller manager"),
							"node_group_operator":      topologyComponentSchema("node group operator"),
							"scheduler":                topologyComponentSchema("scheduler"),
						},
					},
				},
			},
			"version": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Kubernetes version (`MAJOR.MINOR`), the API default when unset. It is checked at plan time against the versions the platform offers. " +
					"Raising it upgrades the cluster in place, lowering it is refused. Removing the attribute keeps the current version",
			},
		},
	}
}

func topologyComponentSchema(name string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Size of the " + name + ", required",
		Attributes: map[string]schema.Attribute{
			"flavor":             schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Flavor of the " + name + ", required"},
			"replication_factor": schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Replication factor of the " + name + ", required"},
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
