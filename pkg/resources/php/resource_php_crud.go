package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

var vhostCleverAppsReg = regexp.MustCompile(`^app-.*\.cleverapps\.io$`)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourcePHP) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "ResourcePHP.Configure()")

	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(provider.Provider)
	if ok {
		r.cc = provider.Client()
		r.org = provider.Organization()
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

	instance := application.LookupInstance(ctx, r.cc, "php", "PHP", resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vhosts := []string{}
	resp.Diagnostics.Append(plan.AdditionalVHosts.ElementsAs(ctx, &vhosts, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment := plan.toEnv(ctx, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createAppReq := application.CreateReq{
		Client:       r.cc,
		Organization: r.org,
		Application: tmp.CreateAppRequest{
			Name:            plan.Name.ValueString(),
			Deploy:          "git",
			Description:     plan.Description.ValueString(),
			InstanceType:    "php",
			InstanceVariant: instance.Variant.ID,
			InstanceVersion: instance.Version,
			BuildFlavor:     plan.BuildFlavor.ValueString(),
			MinFlavor:       plan.SmallestFlavor.ValueString(),
			MaxFlavor:       plan.BiggestFlavor.ValueString(),
			MinInstances:    plan.MinInstanceCount.ValueInt64(),
			MaxInstances:    plan.MaxInstanceCount.ValueInt64(),
			Zone:            plan.Region.ValueString(),
			CancelOnPush:    false,
		},
		Environment: environment,
		VHosts:      vhosts,
	}

	createAppRes := application.CreateApp(ctx, createAppReq, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "BUILD FLAVOR RES"+createAppRes.Application.BuildFlavor.Name, map[string]interface{}{})
	plan.ID = pkg.FromStr(createAppRes.Application.ID)
	plan.DeployURL = pkg.FromStr(createAppRes.Application.DeployURL)
	plan.VHost = pkg.FromStr(createAppRes.Application.Vhosts[0].Fqdn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
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
	state.Name = pkg.FromStr(appPHP.Name)
	state.Description = pkg.FromStr(appPHP.Description)
	state.MinInstanceCount = pkg.FromI(int64(appPHP.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(appPHP.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(appPHP.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appPHP.Instance.MaxFlavor.Name)
	state.Region = pkg.FromStr(appPHP.Zone)
	state.DeployURL = pkg.FromStr(appPHP.DeployURL)

	if appPHP.SeparateBuild {
		state.BuildFlavor = pkg.FromStr(appPHP.BuildFlavor.Name)
	} else {
		state.BuildFlavor = types.StringNull()
	}

	vhosts := pkg.Map(appPHP.Vhosts, func(vhost tmp.Vhost) string {
		return vhost.Fqdn
	})
	hasDefaultVHost := pkg.HasSome(vhosts, func(vhost string) bool {
		return vhostCleverAppsReg.MatchString(vhost)
	})
	if hasDefaultVHost {
		cleverapps := *pkg.First(vhosts, func(vhost string) bool {
			return vhostCleverAppsReg.MatchString(vhost)
		})
		state.VHost = pkg.FromStr(cleverapps)
	} else {
		state.VHost = types.StringNull()
	}

	vhostsWithoutDefault := pkg.Filter(vhosts, func(vhost string) bool {
		ok := vhostCleverAppsReg.MatchString(vhost)
		return !ok
	})
	if len(vhostsWithoutDefault) > 0 {
		state.AdditionalVHosts = pkg.FromListString(vhostsWithoutDefault)
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
