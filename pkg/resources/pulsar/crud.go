package pulsar

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourcePulsar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "ResourcePulsar.Create()", map[string]any{"request": fmt.Sprintf("%+v", req.Plan)})

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
	addonPlan := pkg.LookupProviderPlan(prov, "beta")
	if addonPlan == nil {
		resp.Diagnostics.AddError("failed to find plan", "expect: "+strings.Join(pkg.ProviderPlansAsList(prov), ", "))
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

	pulsarClusterRes := tmp.GetPulsarCluster(ctx, r.Client(), pulsar.ClusterID)
	if pulsarClusterRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar env", pulsarClusterRes.Error().Error())
		return
	}
	pulsarCluster := pulsarClusterRes.Payload()

	read(&plan, pulsar, pulsarCluster)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: handle namespace retention (backlog + offload)
}

// Read resource information
func (r *ResourcePulsar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Pulsar READ", map[string]any{"request": req})

	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	pulsarRes := tmp.GetPulsar(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if pulsarRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar", pulsarRes.Error().Error())
		return
	}
	pulsar := pulsarRes.Payload()

	pulsarClusterRes := tmp.GetPulsarCluster(ctx, r.Client(), pulsar.ClusterID)
	if pulsarClusterRes.HasError() {
		resp.Diagnostics.AddError("failed to get Pulsar env", pulsarClusterRes.Error().Error())
		return
	}
	pulsarCluster := pulsarClusterRes.Payload()

	read(&state, pulsar, pulsarCluster)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func read(state *Pulsar, addon *tmp.Pulsar, cluster *tmp.PulsarCluster) {
	if addon != nil {
		state.Tenant = pkg.FromStr(addon.Tenant)
		state.Namespace = pkg.FromStr(addon.Namespace)
		state.Token = pkg.FromStr(addon.Token)

		// TODO: get addon from ccapi to get the name
		//state.Name = pkg.FromStr(addon.???)
	}

	if cluster == nil {
		return
	}

	state.Region = pkg.FromStr(strings.ToLower(cluster.Zone))

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

// Update resource
func (r *ResourcePulsar) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Pulsar](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r *ResourcePulsar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Pulsar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Pulsar DELETE", map[string]any{"pulsar": state})

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
