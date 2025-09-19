package otoroshi

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

func (r *ResourceOtoroshi) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	otoroshi := helper.PlanFrom[Otoroshi](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()

	provider := pkg.LookupAddonProvider(*addonsProviders, "otoroshi")
	if provider == nil {
		resp.Diagnostics.AddError("Otoroshi provider doesn't exist", fmt.Sprintf("available providers are: %s", strings.Join(pkg.AddonProvidersAsList(*addonsProviders), ", ")))
		return
	}
	plan := provider.Plans[0]

	addonReq := tmp.AddonRequest{
		Name:       otoroshi.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: provider.ID,
		Region:     otoroshi.Region.ValueString(),
	}

	if !otoroshi.Version.IsNull() && !otoroshi.Version.IsUnknown() {
		addonReq.Options = map[string]string{
			"version": otoroshi.Version.ValueString(),
		}
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create Otoroshi add-on", res.Error().Error())
		return
	}
	addonRes := res.Payload()

	otoroshi.ID = pkg.FromStr(addonRes.RealID)
	otoroshi.CreationDate = pkg.FromI(addonRes.CreationDate)

	resp.Diagnostics.Append(resp.State.Set(ctx, otoroshi)...)

	envRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), addonRes.ID)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", envRes.Error().Error())
		return
	}

	envVars := *envRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"env_vars": fmt.Sprintf("%+v", envVars),
	})

	for _, envVar := range envVars {
		switch envVar.Name {
		case "CC_OTOROSHI_API_CLIENT_ID":
			otoroshi.APIClientID = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_API_CLIENT_SECRET":
			otoroshi.APIClientSecret = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_API_URL":
			otoroshi.APIURL = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_INITIAL_ADMIN_LOGIN":
			otoroshi.InitialAdminLogin = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_INITIAL_ADMIN_PASSWORD":
			otoroshi.InitialAdminPassword = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_URL":
			otoroshi.URL = pkg.FromStr(envVar.Value)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, otoroshi)...)
}

func (r *ResourceOtoroshi) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Otoroshi READ", map[string]any{"request": req})

	otoroshi := helper.StateFrom[Otoroshi](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), otoroshi.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otoroshi", addonRes.Error().Error())
		return
	}
	addon := addonRes.Payload()

	addonEnvRes := tmp.GetAddonEnv(ctx, r.Client(), r.Organization(), otoroshi.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	otoroshi.Name = pkg.FromStr(addon.Name)
	otoroshi.Region = pkg.FromStr(addon.Region)
	otoroshi.CreationDate = pkg.FromI(addon.CreationDate)

	envVars := *addonEnvRes.Payload()
	tflog.Debug(ctx, "API response", map[string]any{
		"env_vars": fmt.Sprintf("%+v", envVars),
	})

	for _, envVar := range envVars {
		switch envVar.Name {
		case "CC_OTOROSHI_API_CLIENT_ID":
			otoroshi.APIClientID = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_API_CLIENT_SECRET":
			otoroshi.APIClientSecret = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_API_URL":
			otoroshi.APIURL = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_INITIAL_ADMIN_LOGIN":
			otoroshi.InitialAdminLogin = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_INITIAL_ADMIN_PASSWORD":
			otoroshi.InitialAdminPassword = pkg.FromStr(envVar.Value)
		case "CC_OTOROSHI_URL":
			otoroshi.URL = pkg.FromStr(envVar.Value)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, otoroshi)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ResourceOtoroshi) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[Otoroshi](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[Otoroshi](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID != state.ID {
		resp.Diagnostics.AddError("otoroshi cannot be updated", "mismatched IDs")
		return
	}

	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Otoroshi", addonRes.Error().Error())
		return
	}
	state.Name = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOtoroshi) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var otoroshi Otoroshi

	resp.Diagnostics.Append(req.State.Get(ctx, &otoroshi)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "OTOROSHI DELETE", map[string]any{"otoroshi": otoroshi})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), otoroshi.ID.ValueString())
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
