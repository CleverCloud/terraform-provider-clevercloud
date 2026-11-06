package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// KubernetesV0 is the state before version, topology, tags and persistent
// storage were exposed
type KubernetesV0 struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	KubeConfig           types.String `tfsdk:"kubeconfig"`
	NodeAutoprovisioning types.Bool   `tfsdk:"node_autoprovisioning"`
}

var schemaKubernetesV0 = schema.Schema{
	Version: 0,
	Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true},
		"name":                  schema.StringAttribute{Required: true},
		"kubeconfig":            schema.StringAttribute{Computed: true},
		"node_autoprovisioning": schema.BoolAttribute{Optional: true, Computed: true},
	},
}

// UpgradeState fills the attributes an older state does not hold from the API, as
// it runs even when the refresh is skipped: planning a topology on top of an empty
// state must not read as a change. The provider may not be configured at this
// point and the API may not answer, neither is worth failing every command: the
// attributes then stay null, the next refresh fills them, and ModifyPlan never
// replaces a cluster over a null topology
func (r *ResourceKubernetes) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schemaKubernetesV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, res *resource.UpgradeStateResponse) {
				old := helper.From[KubernetesV0](ctx, *req.State, &res.Diagnostics)
				if res.Diagnostics.HasError() {
					return
				}

				newState := Kubernetes{
					ID:                   old.ID,
					Name:                 old.Name,
					KubeConfig:           old.KubeConfig,
					NodeAutoprovisioning: old.NodeAutoprovisioning,
					PersistentStorage:    types.BoolNull(),
					Tags:                 resources.NullTags,
					Topology:             types.ObjectNull(topologyAttrTypes),
					Version:              types.StringNull(),
				}

				if r.Provider == nil {
					tflog.Warn(ctx, "the provider is not configured, the next refresh will fill the new kubernetes attributes")
					res.Diagnostics.Append(res.State.Set(ctx, newState)...)
					return
				}

				reqCtx, cancel := context.WithTimeout(ctx, nodeAutoprovisioningRequestTimeout)
				defer cancel()

				cluster := tmp.GetKubernetes(reqCtx, r.Client(), r.Organization(), old.ID.ValueString())
				if cluster.HasError() {
					tflog.Warn(ctx, "failed to get the cluster, the next refresh will fill the new kubernetes attributes", map[string]any{"error": cluster.Error().Error()})
					res.Diagnostics.Append(res.State.Set(ctx, newState)...)
					return
				}

				newState.PersistentStorage = pkg.FromBool(persistentStorageFeatureEnabled(cluster.Payload().Features))
				newState.Topology = topologyFromAPI(ctx, cluster.Payload().Topology, &res.Diagnostics)
				newState.Version = pkg.FromStr(cluster.Payload().Version)

				res.Diagnostics.Append(res.State.Set(ctx, newState)...)
			},
		},
	}
}
