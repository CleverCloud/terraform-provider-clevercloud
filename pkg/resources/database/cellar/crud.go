package cellar

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceCellar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	cellar := helper.PlanFrom[Cellar](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}
	addonsProviders := addonsProvidersRes.Payload()

	prov := pkg.LookupAddonProvider(*addonsProviders, "cellar-addon")
	if prov == nil {
		resp.Diagnostics.AddError("failed to find Cellar provider", "")
		return
	}

	plan := prov.FirstPlan()
	if plan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       cellar.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: "cellar-addon",
		Region:     cellar.Region.ValueString(),
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create add-on", res.Error().Error())
		return
	}
	addonRes := res.Payload()
	tflog.Debug(ctx, "get add-on env vars", map[string]any{"cellar": addonRes.RealID})

	cellar.ID = pkg.FromStr(addonRes.RealID)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), addonRes.RealID)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env vars", envRes.Error().Error())
		return
	}
	envVars := envRes.Payload()

	creds := s3.FromEnvVars(*envVars)
	cellar.Host = pkg.FromStr(creds.Host)
	cellar.KeyID = pkg.FromStr(creds.KeyID)
	cellar.KeySecret = pkg.FromStr(creds.KeySecret)

	resp.Diagnostics.Append(resp.State.Set(ctx, cellar)...)
}

// Read resource information
func (r *ResourceCellar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Cellar READ", map[string]any{"request": req})

	cellar := helper.StateFrom[Cellar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), cellar.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Cellar", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), cellar.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}
	addonEnv := addonEnvRes.Payload()

	creds := s3.FromEnvVars(*addonEnv)
	cellar.Name = pkg.FromStr(addon.Name)
	cellar.Region = pkg.FromStr(addon.Region)
	cellar.Host = pkg.FromStr(creds.Host)
	cellar.KeyID = pkg.FromStr(creds.KeyID)
	cellar.KeySecret = pkg.FromStr(creds.KeySecret)

	resp.Diagnostics.Append(resp.State.Set(ctx, cellar)...)
}

// Update resource
func (r *ResourceCellar) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Cellar](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Cellar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID != state.ID {
		resp.Diagnostics.AddError("cellar cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Cellar", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceCellar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := helper.StateFrom[Cellar](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "CELLAR DELETE", map[string]any{"cellar": state})

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Add-on", addonRes.Error().Error())
		return
	}

	// TODO: Use real ID when API supports it
	// res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), cellar.ID.ValueString())
	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonRes.Payload().ID)
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
