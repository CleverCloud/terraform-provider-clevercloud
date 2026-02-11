package otoroshi

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func (r *ResourceOtoroshi) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	state := helper.PlanFrom[Otoroshi](ctx, req.Plan, &resp.Diagnostics)
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
	plan := provider.FirstPlan()
	if plan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       state.Name.ValueString(),
		Plan:       plan.ID,
		ProviderID: provider.ID,
		Region:     state.Region.ValueString(),
	}

	if !state.Version.IsNull() && !state.Version.IsUnknown() {
		addonReq.Options = map[string]string{
			"version": state.Version.ValueString(),
		}
	}

	res := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create Otoroshi add-on", res.Error().Error())
		return
	}
	addonRes := res.Payload()

	state.ID = pkg.FromStr(addonRes.RealID)
	state.CreationDate = pkg.FromI(addonRes.CreationDate)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

	otoroshiRes := tmp.GetOtoroshi(ctx, r.Client(), addonRes.RealID)
	if otoroshiRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otorshi", otoroshiRes.Error().Error())
	} else {
		otoroshi := otoroshiRes.Payload()
		state.APIURL = pkg.FromStr(otoroshi.API.URL)
		state.APIClientID = pkg.FromStr(otoroshi.EnvVars["CC_OTOROSHI_API_CLIENT_ID"])
		state.APIClientSecret = pkg.FromStr(otoroshi.EnvVars["CC_OTOROSHI_API_CLIENT_SECRET"])
		state.InitialAdminLogin = pkg.FromStr(otoroshi.Initialredentials.User)
		state.InitialAdminPassword = pkg.FromStr(otoroshi.Initialredentials.Passsword)
		state.URL = pkg.FromStr(otoroshi.AccessURL)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOtoroshi) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[Otoroshi](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonRes := tmp.GetAddon(ctx, r.Client(), r.Organization(), state.ID.ValueString())
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otoroshi", addonRes.Error().Error())
	} else {
		addon := addonRes.Payload()
		state.Name = pkg.FromStr(addon.Name)
		state.Region = pkg.FromStr(addon.Region)
		state.CreationDate = pkg.FromI(addon.CreationDate)
	}

	otoroshiRes := tmp.GetOtoroshi(ctx, r.Client(), state.ID.ValueString())
	if otoroshiRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otorshi", otoroshiRes.Error().Error())
	} else {
		otoroshi := otoroshiRes.Payload()
		state.APIURL = pkg.FromStr(otoroshi.API.URL)
		state.APIClientID = pkg.FromStr(otoroshi.EnvVars["CC_OTOROSHI_API_CLIENT_ID"])
		state.APIClientSecret = pkg.FromStr(otoroshi.EnvVars["CC_OTOROSHI_API_CLIENT_SECRET"])
		state.InitialAdminLogin = pkg.FromStr(otoroshi.Initialredentials.User)
		state.InitialAdminPassword = pkg.FromStr(otoroshi.Initialredentials.Passsword)
		state.URL = pkg.FromStr(otoroshi.AccessURL)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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
	state.Name = pkg.FromStr(addonRes.Payload().Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOtoroshi) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	otoroshi := helper.StateFrom[Otoroshi](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
