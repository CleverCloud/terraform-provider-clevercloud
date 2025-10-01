package configprovider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Create a new resource
func (r *ResourceConfigProvider) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	addonConfigProvider := helper.PlanFrom[ConfigProvider](ctx, req.Plan, &res.Diagnostics)
	res.Diagnostics.Append(req.Plan.Get(ctx, &addonConfigProvider)...)
	if res.Diagnostics.HasError() {
		return
	}

	addonsProvidersRes := tmp.GetAddonsProviders(ctx, r.Client())
	if addonsProvidersRes.HasError() {
		res.Diagnostics.AddError("failed to get addon providers", addonsProvidersRes.Error().Error())
		return
	}

	addonsProviders := addonsProvidersRes.Payload()
	provider := pkg.LookupAddonProvider(*addonsProviders, "config-provider")

	plan := pkg.LookupProviderPlan(provider, "std")
	if plan == nil {
		res.Diagnostics.AddError("This plan does not exists", "available plans are: "+strings.Join(pkg.ProviderPlansAsList(provider), ", "))
		return
	}

	addonReq := tmp.AddonRequest{
		Name:       addonConfigProvider.Name.ValueString(),
		Region:     addonConfigProvider.Region.ValueString(),
		Plan:       plan.ID,
		ProviderID: "config-provider",
	}

	createAddonRes := tmp.CreateAddon(ctx, r.Client(), r.Organization(), addonReq)
	if createAddonRes.HasError() {
		res.Diagnostics.AddError("failed to create ConfigProvider", createAddonRes.Error().Error())
		return
	}

	addonConfigProvider.ID = pkg.FromStr(createAddonRes.Payload().RealID)
	addonConfigProvider.CreationDate = pkg.FromI(createAddonRes.Payload().CreationDate)

	// Set the state before updating environment variables
	res.Diagnostics.Append(res.State.Set(ctx, addonConfigProvider)...)
	if res.Diagnostics.HasError() {
		return
	}

	envVars := addonConfigProvider.toEnv(ctx, &res.Diagnostics)
	if res.Diagnostics.HasError() {
		return
	}

	// Always initialize an empty slice, even if there are no environment variables
	envVarsArray := []tmp.EnvVar{}

	// Only add environment variables if the map is not empty
	if len(envVars) > 0 {
		// Convert the map to the expected format: []tmp.EnvVar
		for k, v := range envVars {
			envVarsArray = append(envVarsArray, tmp.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	tflog.Debug(ctx, "Setting environment variables on create", map[string]interface{}{"count": len(envVarsArray)})
	envRes := tmp.UpdateConfigProviderEnv(ctx, r.Client(), r.Organization(), addonConfigProvider.ID.ValueString(), envVarsArray)
	if envRes.HasError() {
		res.Diagnostics.AddError("failed to configure application environment", envRes.Error().Error())
		return
	}

	res.Diagnostics.Append(res.State.Set(ctx, addonConfigProvider)...)
	if res.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r *ResourceConfigProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "ConfigProvider READ", map[string]any{"request": req})

	addonConfigProvider := helper.StateFrom[ConfigProvider](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addonEnvRes := tmp.GetConfigProviderEnv(ctx, r.Client(), r.Organization(), addonConfigProvider.ID.ValueString())
	if addonEnvRes.HasError() {
		resp.Diagnostics.AddError("failed to get add-on env", addonEnvRes.Error().Error())
		return
	}

	// Convert the environment variables to a map
	envVars := map[string]string{}
	for _, envVar := range *addonEnvRes.Payload() {
		envVars[envVar.Name] = envVar.Value
	}

	// Update the environment in the state
	addonConfigProvider.fromEnv(ctx, envVars, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the state
	resp.Diagnostics.Append(resp.State.Set(ctx, addonConfigProvider)...)
}

// Update resource
func (r *ResourceConfigProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := helper.PlanFrom[ConfigProvider](ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := helper.StateFrom[ConfigProvider](ctx, req.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.ValueString() != state.ID.ValueString() {
		resp.Diagnostics.AddError("configProvider cannot be updated", "mismatched IDs")
		return
	}

	// Only name can be edited
	addonRes := tmp.UpdateAddon(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), map[string]string{
		"name": plan.Name.ValueString(),
	})
	if addonRes.HasError() {
		resp.Diagnostics.AddError("failed to update ConfigProvider", addonRes.Error().Error())
		return
	}

	tflog.Debug(ctx, "Updating environment variables")
	envVars := plan.toEnv(ctx, &resp.Diagnostics)

	// Always initialize an empty slice, even if there are no environment variables
	envVarsArray := []tmp.EnvVar{}

	// Only add environment variables if the map is not empty
	if len(envVars) > 0 {
		// Convert the map to the expected format: []tmp.EnvVar
		for k, v := range envVars {
			envVarsArray = append(envVarsArray, tmp.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	tflog.Debug(ctx, "Environment variables to update", map[string]interface{}{"count": len(envVarsArray)})
	envRes := tmp.UpdateConfigProviderEnv(ctx, r.Client(), r.Organization(), plan.ID.ValueString(), envVarsArray)
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to configure application environment", envRes.Error().Error())
		return
	}
	state.Name = plan.Name
	state.Environment = plan.Environment

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Delete resource
func (r *ResourceConfigProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	addonConfigProvider := ConfigProvider{}

	diags := req.State.Get(ctx, &addonConfigProvider)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "ConfigProvider DELETE", map[string]any{"configProvider": addonConfigProvider})

	res := tmp.DeleteAddon(ctx, r.Client(), r.Organization(), addonConfigProvider.ID.ValueString())
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
