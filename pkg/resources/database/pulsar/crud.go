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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourcePulsar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Create()")

	plan := helper.PlanFrom[Pulsar](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "addon-pulsar")
	addonPlan := prov.FirstPlan()
	if addonPlan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       plan.Name.ValueString(),
		Plan:       addonPlan.ID,
		ProviderID: "addon-pulsar",
		Region:     plan.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}
	addon := res.Payload()

	plan.ID = pkg.FromStr(addon.RealID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	pulsarRes := tmp.GetPulsar(ctx, r.Client(), r.Organization(), addon.RealID)
	if pulsarRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar", pulsarRes.Error().Error())
		return
	}
	pulsar := pulsarRes.Payload()
	readAddon(&plan, pulsar, &resp.Diagnostics)

	pulsarClusterRes := tmp.GetPulsarCluster(ctx, r.Client(), pulsar.ClusterID)
	if pulsarClusterRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar env", pulsarClusterRes.Error().Error())
		return
	}
	pulsarCluster := pulsarClusterRes.Payload()
	readCluster(&plan, pulsarCluster, &resp.Diagnostics)

	setRetention(ctx, &plan, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: handle namespace retention (offload)
}

// Read resource information
func (r *ResourcePulsar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Read()")

	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()
	readOldAddon(&state, addon, &resp.Diagnostics)

	pulsarRes := tmp.GetPulsar(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if pulsarRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar", pulsarRes.Error().Error())
		return
	}
	pulsar := pulsarRes.Payload()
	readAddon(&state, pulsar, &resp.Diagnostics)

	pulsarClusterRes := tmp.GetPulsarCluster(ctx, r.Client(), pulsar.ClusterID)
	if pulsarClusterRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar env", pulsarClusterRes.Error().Error())
		return
	}
	pulsarCluster := pulsarClusterRes.Payload()
	readCluster(&state, pulsarCluster, &resp.Diagnostics)

	readRetention(ctx, &state, &resp.Diagnostics)

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

func readOldAddon(state *Pulsar, addon *tmp.AddonResponse, diags *diag.Diagnostics) {
	if addon == nil {
		return
	}

	state.Name = pkg.FromStr(addon.Name)
	state.Region = pkg.FromStr(addon.Region)
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
		return // none of the attributs set
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

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("pulsar cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Pulsar", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	// Retention hook
	state.RetentionPeriod = plan.RetentionPeriod
	state.RetentionSize = plan.RetentionSize
	setRetention(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourcePulsar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "ResourcePulsar.Delete()")

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete add-on", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
