package pulsar

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/addon"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourcePulsar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Create()")

	plan := helper.PlanFrom[Pulsar](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonID, createDiags := addon.Create(ctx, r, &plan)
	resp.Diagnostics.Append(createDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set retention after creation (SetFromResponse already read current retention)
	setRetention(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addon.SyncNetworkGroups(ctx, r, addonID, plan.Networkgroups, &plan.Networkgroups, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read resource information
func (r *ResourcePulsar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Read()")

	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonIsDeleted, diags := addon.Read(ctx, r, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if addonIsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func readAddon(state *Pulsar, addon *tmp.Pulsar, diags *diag.Diagnostics) {
	if addon == nil {
		return
	}

	state.Tenant = pkg.FromStr(addon.Tenant)
	state.Namespace = pkg.FromStr(addon.Namespace)
	state.Token = pkg.FromStr(addon.Token)
}

func readCluster(state *Pulsar, cluster *tmp.PulsarCluster, diags *diag.Diagnostics) {
	if cluster == nil {
		return
	}

	if cluster.PulsarTLSPort != 0 {
		state.BinaryURL = pkg.FromStr(fmt.Sprintf("pulsar+ssl://%s:%d", cluster.URL, cluster.PulsarTLSPort))
	} else {
		state.BinaryURL = pkg.FromStr(fmt.Sprintf("pulsar://%s:%d", cluster.URL, cluster.PulsarPort))
	}

	if cluster.WebTLSPort != 0 {
		state.HTTPUrl = pkg.FromStr(fmt.Sprintf("https://%s:%d", cluster.URL, cluster.WebTLSPort))
	} else {
		state.HTTPUrl = pkg.FromStr(fmt.Sprintf("http://%s:%d", cluster.URL, cluster.WebPort))
	}
}

func readRetention(ctx context.Context, state *Pulsar, diags *diag.Diagnostics) {
	period, size := state.RetentionPeriod, state.RetentionSize
	tflog.Debug(ctx, "ReadRetention", map[string]any{"period": period, "size": size})

	admin, err := state.AdminClient()
	if err != nil {
		diags.AddError("failed to create Pulsar admin client", err.Error())
		return
	}

	if !pkg.AtLeastOneSet(period, size) {
		return
	}

	retention, err := admin.Namespaces().GetRetention(state.TenantAndNamespace())
	if err != nil {
		diags.AddError("failed to get Pulsar namespace retention", err.Error())
		return
	}

	pkg.IfIsSetI(period, func(i int64) {
		state.RetentionPeriod = pkg.FromI(retention.RetentionTimeInMinutes)
	})
	pkg.IfIsSetI(size, func(i int64) {
		state.RetentionSize = pkg.FromI(retention.RetentionSizeInMB)
	})
}

// https://pulsar.apache.org/docs/next/cookbooks-retention-expiry/#retention-policies
func setRetention(ctx context.Context, plan *Pulsar, diags *diag.Diagnostics) {
	size := plan.RetentionSize
	period := plan.RetentionPeriod

	tflog.Debug(ctx, "SetRetention", map[string]any{"period": period, "size": size, "tenantNs": plan.TenantAndNamespace()})

	if !pkg.AtLeastOneSet(size, period) {
		return // none of the attributes set
	}

	admin, err := plan.AdminClient()
	if err != nil {
		diags.AddError("failed to create Pulsar admin client", err.Error())
		return
	}

	// we know at least 1 param is set, so we can relax the other to prevent invalid cases
	policy := utils.RetentionPolicies{RetentionTimeInMinutes: -1, RetentionSizeInMB: -1}

	pkg.IfIsSetI(period, func(i int64) {
		policy.RetentionTimeInMinutes = int(period.ValueInt64())
	})

	pkg.IfIsSetI(size, func(i int64) {
		policy.RetentionSizeInMB = size.ValueInt64()
	})

	err = admin.Namespaces().SetRetention(plan.TenantAndNamespace(), policy)
	if err != nil {
		diags.AddError("failed to set Pulsar retention", err.Error())
		return
	}
}

// Update resource
func (r *ResourcePulsar) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Update()")

	plan := helper.PlanFrom[Pulsar](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonID, updateDiags := addon.Update(ctx, r, &plan, &state)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save plan retention values before SetFromResponse overwrites them
	retentionPeriod := plan.RetentionPeriod
	retentionSize := plan.RetentionSize

	// Fill Computed fields (tenant, namespace, URLs, token...) from API
	plan.SetFromResponse(ctx, r.Client(), r.Organization(), addonID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore plan retention values and apply them
	plan.RetentionPeriod = retentionPeriod
	plan.RetentionSize = retentionSize
	setRetention(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addon.SyncNetworkGroups(ctx, r, addonID, plan.Networkgroups, &plan.Networkgroups, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete resource
func (r *ResourcePulsar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Delete()")

	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(addon.Delete(ctx, r, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
