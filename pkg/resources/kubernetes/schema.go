package kubernetes

import (
	"context"
	_ "embed"
	"regexp"

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
	Region               types.String `tfsdk:"region"`
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
		Version:             1,
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
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					regionOnUnknownZone{},
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Zone the cluster runs in, by name (`par`). Without it the API picks the only zone the organisation may use, else the platform's default zone, and the attribute reads the zone it picked. " +
					"It is checked at plan time against the zones the organisation may use. Changing it replaces the cluster. Setting it on a cluster whose zone the API does not name fails the plan, taint the resource to replace the cluster",
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
						Validators:          flavorValidators,
						MarkdownDescription: "Flavor of the control plane in uppercase, required unless `type` is `DISTRIBUTED`. The accepted flavors depend on the topology and are checked at plan time",
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
					"Changing it moves the cluster to that version in place, in either direction. Removing the attribute keeps the current version",
			},
		},
	}
}

// regionOnUnknownZone refuses a region on a cluster whose zone the state does not
// hold, which happens when the API does not name it: the cluster may already run
// there, and replacing it would destroy it for nothing
type regionOnUnknownZone struct{}

func (regionOnUnknownZone) Description(context.Context) string {
	return "refuses a region on a cluster whose zone is unknown"
}

func (m regionOnUnknownZone) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (regionOnUnknownZone) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || !req.StateValue.IsNull() || req.ConfigValue.IsNull() {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "the zone of this cluster is unknown",
		"The API does not name the zone this cluster runs in, so setting region could destroy a cluster that already runs there. "+
			"Remove region, or taint this resource to replace the cluster on purpose")
}

// flavorValidators refuses a flavor in any other case. The API takes them all and
// answers in uppercase, so a lowercase value would never match the state and would
// replace the cluster on every plan. The accepted flavors are not listed here: the
// platform decides them, and ModifyPlan checks them against it
var flavorValidators = []validator.String{
	stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9A-Z]+$`), "must be uppercase, for example XS"),
}

func topologyComponentSchema(name string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Size of the " + name + ", required",
		Attributes: map[string]schema.Attribute{
			"flavor":             schema.StringAttribute{Optional: true, Computed: true, Validators: flavorValidators, MarkdownDescription: "Flavor of the " + name + " in uppercase, required"},
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
