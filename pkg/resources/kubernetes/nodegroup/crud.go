package nodegroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.dev/sdk/models"
)

func (r *ResourceKubernetesNodegroup) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	plan := helper.PlanFrom[KubernetesNodegroup](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Create Kubernetes cluster
	createRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(plan.KubernetesID.ValueString()).
		NodeGroups().
		Createnodegroup(ctx, &models.NodeGroupCreationPayload{
			Name:            plan.Name.ValueString(),
			Flavor:          models.NodeFlavor(plan.Flavor.ValueString()),
			TargetNodeCount: int(plan.Size.ValueInt64()),
			MinNodeCount:    int(plan.Size.ValueInt64()),
			MaxNodeCount:    int(plan.Size.ValueInt64()),
			Labels:          &models.MapString{},
		})
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
		Getnodegroup(ctx)
	if ngRes.HasError() {
		res.Diagnostics.AddError("failed to get nodegroup", ngRes.Error().Error())
	}
	nodegroup := ngRes.Payload()

	state.ID = identity.ID
	state.Name = pkg.FromStr(nodegroup.Name)
	state.Flavor = pkg.FromStr(string(nodegroup.Flavor))
	state.Size = pkg.FromI(nodegroup.TargetNodeCount)

	res.Diagnostics.Append(res.State.Set(ctx, state)...)
}

func (r *ResourceKubernetesNodegroup) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	identity := helper.From[KubernetesNodegroupIdentity](ctx, req.Identity, &res.Diagnostics)
	plan := helper.From[KubernetesNodegroup](ctx, req.Plan, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	updateRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(plan.KubernetesID.ValueString()).NodeGroups().
		Nodegroupid(identity.ID.ValueString()).Updatenodegroup(ctx, &models.NodeGroupPatchPayload{
		Flavor:          models.NodeFlavor(plan.Flavor.ValueString()),
		Name:            plan.Name.ValueString(),
		TargetNodeCount: int(plan.Size.ValueInt64()),
		MinNodeCount:    int(plan.Size.ValueInt64()),
		MaxNodeCount:    int(plan.Size.ValueInt64()),
	})
	/*tmp.UpdateNodeGroup(
	ctx,
	r.Client(),
	r.Organization(),
	plan.KubernetesID.ValueString(),
	identity.ID.ValueString(),
	updateReq)*/
	if updateRes.HasError() {
		res.Diagnostics.AddError("failed to update nodegroup", updateRes.Error().Error())
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

	deleteRes := r.SDK.V4().Kubernetes().
		Organisations().Ownerid(r.Organization()).
		Clusters().Clusterid(state.KubernetesID.ValueString()).
		NodeGroups().Nodegroupid(identity.ID.ValueString()).
		Deletenodegroup(ctx)
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
