package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourcePHP) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "ResourcePHP.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*Provider)
	if ok {
		r.cc = provider.cc
		r.org = provider.Organisation
	}

	tflog.Info(ctx, "AFTER CONFIGURED", map[string]interface{}{"cc": r.cc == nil, "org": r.org})
}

// Create a new resource
func (r *ResourcePHP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := PHP{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GET variants
	var version string
	var variantID string
	productRes := tmp.GetProductInstance(ctx, r.cc)
	if productRes.HasError() {
		resp.Diagnostics.AddError("failed to get variant", productRes.Error().Error())
		return
	}
	for _, product := range *productRes.Payload() {
		if product.Type != "php" || product.Name != "PHP" {
			continue
		}

		version = product.Version
		variantID = product.Variant.ID
		break
	}
	if version == "" || variantID == "" {
		resp.Diagnostics.AddError("failed to get variant", "there id no product matching 'node'")
		return
	}

	tflog.Info(ctx, "BUILD FLAVOR "+plan.BuildFlavor.String())
	createAppReq := tmp.CreateAppRequest{
		Name:            plan.Name.ValueString(),
		Deploy:          "git",
		Description:     plan.Description.ValueString(),
		InstanceType:    "php",
		InstanceVariant: variantID,
		InstanceVersion: version,
		BuildFlavor:     plan.BuildFlavor.ValueString(),
		MinFlavor:       plan.SmallestFlavor.ValueString(),
		MaxFlavor:       plan.BiggestFlavor.ValueString(),
		MinInstances:    plan.MinInstanceCount.ValueInt64(),
		MaxInstances:    plan.MaxInstanceCount.ValueInt64(),
		Zone:            plan.Region.ValueString(),
		CancelOnPush:    false,
	}

	res := tmp.CreateApp(ctx, r.cc, r.org, createAppReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create app", res.Error().Error())
		return
	}

	appRes := res.Payload()
	tflog.Info(ctx, "BUILD FLAVOR RES"+appRes.BuildFlavor.Name, map[string]interface{}{})
	plan.ID = fromStr(appRes.ID)
	plan.DeployURL = fromStr(appRes.DeployURL)
	plan.VHost = fromStr(appRes.Vhosts[0].Fqdn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envRes := tmp.UpdateAppEnv(ctx, r.cc, r.org, appRes.ID, plan.toEnv())
	if envRes.HasError() {
		resp.Diagnostics.AddError("failed to configure application", envRes.Error().Error())
	}

	vhosts := []string{}
	resp.Diagnostics.Append(plan.AdditionalVHosts.ElementsAs(ctx, &vhosts, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vhostsWithoutDefault := Filter(vhosts, func(vhost string) bool {
		ok := VhostCleverAppsReg.MatchString(vhost)
		return !ok
	})

	for _, vhost := range vhostsWithoutDefault {
		addVhostRes := tmp.AddAppVHost(ctx, r.cc, r.org, appRes.ID, vhost)
		if addVhostRes.HasError() {
			resp.Diagnostics.AddError("failed to add additional vhost", addVhostRes.Error().Error())
		}
	}
}

// Read resource information
func (r *ResourcePHP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PHP

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appRes := tmp.GetApp(ctx, r.cc, r.org, state.ID.ValueString())
	if appRes.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if appRes.HasError() {
		resp.Diagnostics.AddError("failed to get app", appRes.Error().Error())
	}

	appPHP := appRes.Payload()
	state.Name = fromStr(appPHP.Name)
	state.Description = fromStr(appPHP.Description)
	state.MinInstanceCount = fromI(int64(appPHP.Instance.MinInstances))
	state.MaxInstanceCount = fromI(int64(appPHP.Instance.MaxInstances))
	state.SmallestFlavor = fromStr(appPHP.Instance.MinFlavor.Name)
	state.BiggestFlavor = fromStr(appPHP.Instance.MaxFlavor.Name)
	state.Region = fromStr(appPHP.Zone)
	state.DeployURL = fromStr(appPHP.DeployURL)

	if appPHP.SeparateBuild {
		state.BuildFlavor = fromStr(appPHP.BuildFlavor.Name)
	} else {
		state.BuildFlavor = types.StringNull()
	}

	vhosts := Map(appPHP.Vhosts, func(vhost tmp.Vhost) string {
		return vhost.Fqdn
	})
	hasDefaultVHost := HasSome(vhosts, func(vhost string) bool {
		return VhostCleverAppsReg.MatchString(vhost)
	})
	if hasDefaultVHost {
		cleverapps := *First(vhosts, func(vhost string) bool {
			return VhostCleverAppsReg.MatchString(vhost)
		})
		state.VHost = fromStr(cleverapps)
	} else {
		state.VHost = types.StringNull()
	}

	vhostsWithoutDefault := Filter(vhosts, func(vhost string) bool {
		ok := VhostCleverAppsReg.MatchString(vhost)
		return !ok
	})
	if len(vhostsWithoutDefault) > 0 {
		state.AdditionalVHosts = fromListString(vhostsWithoutDefault)
	} else {
		state.AdditionalVHosts = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourcePHP) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourcePHP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PHP

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "PHP DELETE", map[string]interface{}{"state": state})

	res := tmp.DeleteApp(ctx, r.cc, r.org, state.ID.ValueString())
	if res.IsNotFoundError() {
		resp.State.RemoveResource(ctx)
		return
	}
	if res.HasError() {
		resp.Diagnostics.AddError("failed to delete app", res.Error().Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// Import resource
func (r *ResourcePHP) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
