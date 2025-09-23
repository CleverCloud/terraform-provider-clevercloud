package matomo

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
func (r *ResourceMatomo) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	appMatomo := helper.PlanFrom[Matomo](ctx, req.Plan, &res.Diagnostics)
	res.Diagnostics.Append(req.Plan.Get(ctx, &appMatomo)...)
	if res.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		res.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	provider := pkg.LookupAddonProvider(*addonsProviders, "addon-matomo")

	// For now, only the "beta" plan is available
	plan := pkg.LookupProviderPlan(provider, "beta") // appMatomo.Plan.ValueString())
	if plan == nil {
		res.Diagnostics.AddError("This plan does not exists", "available plans are: " +strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       appMatomo.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "addon-matomo",
		Region:     appMatomo.Region.ValueString(),
	}

	createAddonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create Matomo", createAddonRes.Error().Error())
		return
	}

	appMatomo.ID = pkg.FromStr(createAddonRes.Payload().RealID)
	appMatomo.CreationDate = pkg.FromI(createAddonRes.Payload().CreationDate)

	res.Diagnostics.Append(res.State.Set(ctx, appMatomo)...)

	appMatomoEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), appMatomo.ID.ValueString())
	if appMatomoEnvRes.HasError() {
		res.Diagnostics.AddError("failed to get Matomo connection infos", appMatomoEnvRes.Error().Error())
		return
	}

	appMatomoEnv := *appMatomoEnvRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"payload": fmt.Sprintf("%+v", appMatomoEnv),
	})

	hostEnvVar := pkg.First(appMatomoEnv, func(v tmp.EnvVar) bool {
		return v.Name == "MATOMO_URL"
	})
	if hostEnvVar == nil {
		res.Diagnostics.AddError("cannot get Matomo infos", "missing MATOMO_URL env var on created addon")
		return
	}

	appMatomo.Host = pkg.FromStr(hostEnvVar.Value)

	res.Diagnostics.Append(res.State.Set(ctx, appMatomo)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceMatomo) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Matomo READ", map[string]any{"request": req})

	var appMatomo Matomo
	diags := req.State.Get(ctx, &appMatomo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO

	diags = resp.State.Set(ctx, appMatomo)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceMatomo) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Matomo](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Matomo](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("matomo cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Matomo", addonRes.Error().Error())
		return
	}
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceMatomo) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	appMatomo := Matomo{}

	diags := req.State.Get(ctx, &appMatomo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Matomo DELETE", map[string]any{"matomo": appMatomo})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), appMatomo.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete addon", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
