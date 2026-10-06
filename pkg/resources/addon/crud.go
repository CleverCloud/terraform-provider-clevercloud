package addon

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
func (r *ResourceAddon) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ad := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	provider := pkg.LookupAddonProvider(*addonsProviders, ad.ThirdPartyProvider.ValueString())
	if provider == nil {
		resp.Diagnostics.AddError("This provider doesn't exist", fmt.Sprintf("available providers are: %s", strings.Join(pkg.AddonProvidersAsList(*addonsProviders), ", ")))
		return
	}

	plan := pkg.LookupProviderPlan(provider, ad.Plan.ValueString())
	if plan == nil {
		resp.Diagnostics.AddError("This plan doesn't exist", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       ad.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: provider.ID,
		Region:     ad.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}

	ad.ID = pkg.FromStr(res.Payload().ID)
	ad.CreationDate = pkg.FromI(res.Payload().CreationDate)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), res.Payload().ID)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", res.Error().Error())
		return
	}

	ad.FromEnv(ctx, *envRes.Payload(), &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, ad)...)
}

// Read resource information
func (r *ResourceAddon) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Add-on READ", map[string]any{"request": req})

	var ad Addon
	diags := req.State.Get(ctx, &ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if ad.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonRes.IsNotFoundError() {
		req.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on", addonRes.Error().Error())
		return
	}

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	ad.
		FromAddon(ctx, addonRes.Payload(), &resp.Diagnostics).
		FromEnv(ctx, *addonEnvRes.Payload(), &resp.Diagnostics)

	diags = resp.State.Set(ctx, ad)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r *ResourceAddon) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Addon](ctx, req.Plan, &resp.Diagnostics)
	state := helper.StateFrom[Addon](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("addon cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Addon", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceAddon) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var ad Addon

	diags := req.State.Get(ctx, &ad)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Add-on DELETE", map[string]any{"addon": ad})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), ad.ID.ValueString())
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
