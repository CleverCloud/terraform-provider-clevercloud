package nodegroup

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/models"
)

func (r *ResourceKubernetesNodegroup) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	plan := helper.PlanFrom[KubernetesNodegroup](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Create Kubernetes cluster
	createRes := r.SDK.V4().
		Kubernetes().
		Organisations().
		Ownerid(r.Organization()).
		Clusters().
		Clusterid(plan.KubernetesID.ValueString()).
		NodeGroups().
		Createkubernetesnodegroup(ctx, &models.NodeGroupCreationPayload{
			Name:            plan.Name.ValueString(),
			Flavor:          models.NodeFlavor(plan.Flavor.ValueString()),
			TargetNodeCount: int(plan.Size.ValueInt64()),
			MinNodeCount:    pkg.AsPointer(plan.Size),
			MaxNodeCount:    pkg.AsPointer(plan.Size),
			Labels:          &models.MapLabelkeyLabelvalue{},
		})
	if createRes.HasError() {
		var apiError client.APIError
		if errors.As(createRes.Error(), &apiError) {
			tflog.Debug(ctx, "API error", map[string]any{
				"context": apiError.Context,
				"code":    apiError.Code,
			}) // TODO: find a better way to have relevant infos
		}

		res.Diagnostics.AddError("failed to create kubernetes nodegroup", createRes.Error().Error())
		return
	}
	nodegroup := createRes.Payload()

	identity := KubernetesNodegroupIdentity{ID: pkg.FromStr(nodegroup.ID)}
	res.Diagnostics.Append(res.Identity.Set(ctx, identity)...)
	if res.Diagnostics.HasError() {
		return
	}

	state := KubernetesNodegroup{
		ID:           pkg.FromStr(nodegroup.ID),
		KubernetesID: pkg.FromStr(nodegroup.ClusterID),
		Name:         pkg.FromStr(nodegroup.Name),
		Flavor:       pkg.FromStr(string(nodegroup.Flavor)),
		Size:         pkg.FromI(nodegroup.TargetNodeCount),
	}

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceKubernetesNodegroup) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	identity := helper.From[KubernetesNodegroupIdentity](ctx, req.Identity, &res.Diagnostics)
	state := helper.From[KubernetesNodegroup](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	if identity.ID.ValueString() == "" {
		res.State.RemoveResource(ctx)
		return
	}

	ngRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(state.KubernetesID.ValueString()).
		NodeGroups().Nodegroupid(identity.ID.ValueString()).
		Getkubernetesnodegroup(ctx)
	if ngRes.IsNotFoundError() {
		res.State.RemoveResource(ctx)
		return
	}
	if ngRes.HasError() {
		res.Diagnostics.AddError("failed to get nodegroup", ngRes.Error().Error())
		return
	}
	nodegroup := ngRes.Payload()
	if nodegroup.Status == models.NodeGroupStatusTypeDELETED {
		res.State.RemoveResource(ctx)
		return
	}
	if !canManageNodegroup(nodegroup, &res.Diagnostics) {
		return
	}

	state.ID = identity.ID
	state.Name = pkg.FromStr(nodegroup.Name)
	state.Flavor = pkg.FromStr(string(nodegroup.Flavor))
	state.Size = pkg.FromI(nodegroup.TargetNodeCount)

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceKubernetesNodegroup) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// Keep the prior state until a mutation succeeds.
	res.State = req.State
	identity := helper.From[KubernetesNodegroupIdentity](ctx, req.Identity, &res.Diagnostics)
	plan := helper.From[KubernetesNodegroup](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Read may have been skipped with -refresh=false, and labels may have changed
	// since the plan. Check the current group before sending any mutation.
	ngRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(plan.KubernetesID.ValueString()).
		NodeGroups().Nodegroupid(identity.ID.ValueString()).
		Getkubernetesnodegroup(ctx)
	if ngRes.HasError() && !ngRes.IsNotFoundError() {
		res.Diagnostics.AddError("failed to get nodegroup", ngRes.Error().Error())
		return
	}
	// Deleted groups may remain in the API with their old Karpenter labels.
	// Update cannot remove state; a new plan must reconcile the deletion.
	if ngRes.IsNotFoundError() || ngRes.Payload().Status == models.NodeGroupStatusTypeDELETED {
		res.Diagnostics.AddError("nodegroup no longer exists", fmt.Sprintf(
			"Node group %q no longer exists. Run a new Terraform plan with refresh enabled to reconcile its deletion before applying changes.",
			identity.ID.ValueString(),
		))
		return
	}
	if !canManageNodegroup(ngRes.Payload(), &res.Diagnostics) {
		return
	}

	updateRes := r.SDK.V4().
		Kubernetes().
		Organisations().
		Ownerid(r.Organization()).
		Clusters().
		Clusterid(plan.KubernetesID.ValueString()).NodeGroups().
		Nodegroupid(identity.ID.ValueString()).
		Updatekubernetesnodegroup(ctx, &models.NodeGroupPatchPayload{
			Name:            plan.Name.ValueString(),
			TargetNodeCount: int(plan.Size.ValueInt64()),
			MinNodeCount:    pkg.AsPointer(plan.Size),
			MaxNodeCount:    pkg.AsPointer(plan.Size),
		})
	if updateRes.HasError() {
		res.Diagnostics.AddError("failed to update nodegroup", updateRes.Error().Error())
		return
	}
	nodegroup := updateRes.Payload()

	state := KubernetesNodegroup{
		ID:           identity.ID,
		KubernetesID: plan.KubernetesID,
		Name:         pkg.FromStr(nodegroup.Name),
		Flavor:       pkg.FromStr(string(nodegroup.Flavor)),
		Size:         pkg.FromI(nodegroup.TargetNodeCount),
	}

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceKubernetesNodegroup) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	identity := helper.From[KubernetesNodegroupIdentity](ctx, req.Identity, &res.Diagnostics)
	state := helper.From[KubernetesNodegroup](ctx, req.State, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	ngRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(state.KubernetesID.ValueString()).
		NodeGroups().Nodegroupid(identity.ID.ValueString()).
		Getkubernetesnodegroup(ctx)
	if ngRes.IsNotFoundError() {
		res.State.RemoveResource(ctx)
		return
	}
	if ngRes.HasError() {
		res.Diagnostics.AddError("failed to get nodegroup", ngRes.Error().Error())
		return
	}
	nodegroup := ngRes.Payload()
	// The API also retains deleted groups as HTTP 200 records.
	if nodegroup.Status == models.NodeGroupStatusTypeDELETED {
		res.State.RemoveResource(ctx)
		return
	}
	if !canManageNodegroup(nodegroup, &res.Diagnostics) {
		return
	}

	deleteRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(state.KubernetesID.ValueString()).
		NodeGroups().Nodegroupid(identity.ID.ValueString()).
		Deletekubernetesnodegroup(ctx)
	if deleteRes.IsNotFoundError() {
		res.State.RemoveResource(ctx)
		return
	}
	if deleteRes.HasError() {
		res.Diagnostics.AddError("failed to delete nodegroup", deleteRes.Error().Error())
		return
	}

	res.State.RemoveResource(ctx)
}

// The v4 API exposes node labels, but not the Kubernetes ownership metadata.
// A nonempty NodePool label is the marker Karpenter propagates to its groups.
// This guard intentionally leaves fixed groups on the same cluster manageable.
func canManageNodegroup(nodegroup *models.NodeGroup, diags *diag.Diagnostics) bool {
	const nodePoolLabel = "karpenter.sh/nodepool"
	nodePool := nodegroup.Labels[nodePoolLabel]
	if nodePool == "" {
		return true
	}

	diags.AddError("nodegroup carries Karpenter labels", fmt.Sprintf(
		"Node group %q carries the label %s=%q. Leave this group under Karpenter's control; "+
			"clevercloud_kubernetes_nodegroup cannot manage it. If it is already in Terraform state, "+
			"remove its resource configuration and use a removed block with destroy = false or terraform state rm to forget it without deleting it.",
		nodegroup.ID, nodePoolLabel, nodePool,
	))
	return false
}
