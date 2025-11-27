package nodegroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceKubernetesNodegroup) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	res.Diagnostics.AddWarning(
		"Kubernetes product support is in beta",
		"It can break at any time, use it at your own risks",
	)

	plan := helper.PlanFrom[KubernetesNodegroup](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Create Kubernetes cluster
	createReq := tmp.NodeGroupCreationPayload{
		Name:            plan.Name.ValueString(),
		Flavor:          plan.Flavor.ValueString(),
		TargetNodeCount: int32(plan.Size.ValueInt64()),
		MinNodeCount:    int32(plan.Size.ValueInt64()),
		MaxNodeCount:    int32(plan.Size.ValueInt64()),
		Labels:          map[string]string{},
	}

	createRes := tmp.CreateNodeGroup(
		ctx,
		r.Client(),
		r.Organization(),
		plan.KubernetesID.ValueString(),
		createReq,
	)
	if createRes.HasError() {
		res.Diagnostics.AddError("failed to create kubernetes nodegroup", createRes.Error().Error())
		return
	}
	nodegroup := createRes.Payload()

	identity := KubernetesNodegroupIdentity{ID: pkg.FromStr(nodegroup.ID)}
	res.Diagnostics.Append(res.Identity.Set(ctx, identity)...)

	state := KubernetesNodegroup{
		ID:           pkg.FromStr(nodegroup.ID),
		KubernetesID: pkg.FromStr(nodegroup.ClusterID),
		Name:         pkg.FromStr(nodegroup.Name),
		Flavor:       pkg.FromStr(nodegroup.Flavor),
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

	ngRes := tmp.GetNodeGroup(
		ctx,
		r.Client(),
		r.Organization(),
		state.KubernetesID.ValueString(),
		identity.ID.ValueString(),
	)
	if ngRes.HasError() {
		res.Diagnostics.AddError("failed to get nodegroup", ngRes.Error().Error())
	}
	nodegroup := ngRes.Payload()

	state.ID = identity.ID
	state.Name = pkg.FromStr(nodegroup.Name)
	state.Flavor = pkg.FromStr(nodegroup.Flavor)
	state.Size = pkg.FromI(nodegroup.TargetNodeCount)

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceKubernetesNodegroup) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	identity := helper.From[KubernetesNodegroupIdentity](ctx, req.Identity, &res.Diagnostics)
	plan := helper.From[KubernetesNodegroup](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	updateReq := tmp.NodeGroupPatchPayload{
		Flavor:          *plan.Flavor.ValueStringPointer(),
		Name:            plan.Name.ValueString(),
		TargetNodeCount: int32(plan.Size.ValueInt64()),
		MinNodeCount:    int32(plan.Size.ValueInt64()),
		MaxNodeCount:    int32(plan.Size.ValueInt64()),
	}

	updateRes := tmp.UpdateNodeGroup(
		ctx,
		r.Client(),
		r.Organization(),
		plan.KubernetesID.ValueString(),
		identity.ID.ValueString(),
		updateReq)
	if updateRes.HasError() {
		res.Diagnostics.AddError("failed to update nodegroup", updateRes.Error().Error())
	}
	nodegroup := updateRes.Payload()

	state := KubernetesNodegroup{
		ID:           identity.ID,
		KubernetesID: plan.KubernetesID,
		Name:         pkg.FromStr(nodegroup.Name),
		Flavor:       pkg.FromStr(nodegroup.Flavor),
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

	deleteRes := tmp.DeleteNodeGroup(
		ctx,
		r.Client(),
		r.Organization(),
		state.KubernetesID.ValueString(),
		identity.ID.ValueString(),
	)
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
