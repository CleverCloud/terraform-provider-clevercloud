package otoroshi

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/addons"
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
	plan := pkg.FirstPlan(provider)
	if plan == nil {
		resp.Diagnostics.AddError("at least 1 plan for addon is required", "no plans")
		return
	}

	addonReq := models.WannabeAddonProvision{
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

	res := resources.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create Otoroshi add-on", res.Error().Error())
		return
	}
	addonRes := res.Payload()

	state.ID = pkg.FromStr(pkg.Deref(addonRes.RealID))
	state.CreationDate = pkg.FromI(pkg.Deref(addonRes.CreationDate))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

	otoroshiRes := r.SDK.V4().AddonProviders().AddonOtoroshi().Addons().Otoroshiid(pkg.Deref(addonRes.RealID)).Getotoroshi(ctx)
	if otoroshiRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otorshi", otoroshiRes.Error().Error())
	} else {
		otoroshi := otoroshiRes.Payload()
		if otoroshi.API.URL != "" {
			state.APIURL = pkg.FromStr(otoroshi.API.URL)
			state.APIClientID = pkg.FromStr(otoroshi.API.User)
			state.APIClientSecret = pkg.FromStr(otoroshi.API.Secret)
		}
		state.InitialAdminLogin = pkg.FromStr(otoroshi.InitialCredentials.User)
		state.InitialAdminPassword = pkg.FromStr(otoroshi.InitialCredentials.Password)
		state.URL = pkg.FromStr(otoroshi.AccessURL)

		// Networkgroups are piloted on the Java application backing the add-on
		application.SyncNetworkGroups(ctx, r, otoroshi.Resources.Entrypoint, state.Networkgroups, &resp.Diagnostics)
	}

	state.Tags = resources.SyncAddonTags(ctx, r, addonRes.ID, state.Tags, resources.NullTags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOtoroshi) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := helper.StateFrom[Otoroshi](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(state.ID.ValueString()).Getaddonbyorgaandaddonid(ctx)
	if addonRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	} else if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otoroshi", addonRes.Error().Error())
	} else {
		addon := addonRes.Payload()
		state.Name = pkg.FromStr(pkg.Deref(addon.Name))
		state.Region = pkg.FromStr(addon.Region)
		state.CreationDate = pkg.FromI(pkg.Deref(addon.CreationDate))
	}

	otoroshiRes := r.SDK.V4().AddonProviders().AddonOtoroshi().Addons().Otoroshiid(state.ID.ValueString()).Getotoroshi(ctx)
	if otoroshiRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otorshi", otoroshiRes.Error().Error())
	} else {
		otoroshi := otoroshiRes.Payload()
		if otoroshi.API.URL != "" {
			state.APIURL = pkg.FromStr(otoroshi.API.URL)
			state.APIClientID = pkg.FromStr(otoroshi.API.User)
			state.APIClientSecret = pkg.FromStr(otoroshi.API.Secret)
		}
		state.InitialAdminLogin = pkg.FromStr(otoroshi.InitialCredentials.User)
		state.InitialAdminPassword = pkg.FromStr(otoroshi.InitialCredentials.Password)
		state.URL = pkg.FromStr(otoroshi.AccessURL)

		// Networkgroups are piloted on the Java application backing the add-on
		state.Networkgroups = resources.ReadNetworkGroups(ctx, r, otoroshi.Resources.Entrypoint, &resp.Diagnostics)
	}

	state.Tags = resources.ReadAddonTags(ctx, r, state.ID.ValueString(), state.Tags, &resp.Diagnostics)

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

	addonRes := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(plan.ID.ValueString()).Updateaddoninfo(ctx, addons.UpdateaddoninfoBody{
		Name: plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update Otoroshi", addonRes.Error().Error())
		return
	}
	state.Name = pkg.FromStr(pkg.Deref(addonRes.Payload().Name))

	// Networkgroups are piloted on the Java application backing the add-on
	otoroshiRes := r.SDK.V4().AddonProviders().AddonOtoroshi().Addons().Otoroshiid(plan.ID.ValueString()).Getotoroshi(ctx)
	if otoroshiRes.HasError() {
		resp.Diagnostics.AddError("failed to get Otoroshi", otoroshiRes.Error().Error())
	} else {
		application.SyncNetworkGroups(ctx, r, otoroshiRes.Payload().Resources.Entrypoint, plan.Networkgroups, &resp.Diagnostics)
		state.Networkgroups = plan.Networkgroups
	}

	if !plan.Tags.Equal(state.Tags) {
		state.Tags = resources.SyncAddonTags(ctx, r, state.ID.ValueString(), plan.Tags, state.Tags, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ResourceOtoroshi) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	otoroshi := helper.StateFrom[Otoroshi](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := r.SDK.V2().Organisations().ID(r.Organization()).Addons().Addonid(otoroshi.ID.ValueString()).Deprovisionaddonbyorgaandaddonid(ctx)
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
