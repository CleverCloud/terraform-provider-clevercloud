package java

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/application"
	"go.clever-cloud.com/terraform-provider/pkg/provider"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceJava) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "ResourceJava.Configure()")

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
func (r *ResourceJava) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := Java{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance := application.LookupInstance(ctx, r.cc, "java", r.toProductName(), resp.Diagnostics)
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
			InstanceType:    instance.Type,
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
		Deployment:  plan.toDeployment(),
	}

	createAppRes, diags := application.CreateApp(ctx, createAppReq)
	resp.Diagnostics.Append(diags...)
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
func (r *ResourceJava) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Java

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

	appJava := appRes.Payload()
	state.Name = pkg.FromStr(appJava.Name)
	state.Description = pkg.FromStr(appJava.Description)
	state.MinInstanceCount = pkg.FromI(int64(appJava.Instance.MinInstances))
	state.MaxInstanceCount = pkg.FromI(int64(appJava.Instance.MaxInstances))
	state.SmallestFlavor = pkg.FromStr(appJava.Instance.MinFlavor.Name)
	state.BiggestFlavor = pkg.FromStr(appJava.Instance.MaxFlavor.Name)
	state.Region = pkg.FromStr(appJava.Zone)
	state.DeployURL = pkg.FromStr(appJava.DeployURL)

	if appJava.SeparateBuild {
		state.BuildFlavor = pkg.FromStr(appJava.BuildFlavor.Name)
	} else {
		state.BuildFlavor = types.StringNull()
	}

	vhosts := pkg.Map(appJava.Vhosts, func(vhost tmp.Vhost) string {
		return vhost.Fqdn
	})
	hasDefaultVHost := pkg.HasSome(vhosts, func(vhost string) bool {
		return pkg.VhostCleverAppsRegExp.MatchString(vhost)
	})
	if hasDefaultVHost {
		cleverapps := *pkg.First(vhosts, func(vhost string) bool {
			return pkg.VhostCleverAppsRegExp.MatchString(vhost)
		})
		state.VHost = pkg.FromStr(cleverapps)
	} else {
		state.VHost = types.StringNull()
	}

	vhostsWithoutDefault := pkg.Filter(vhosts, func(vhost string) bool {
		ok := pkg.VhostCleverAppsRegExp.MatchString(vhost)
		return !ok
	})
	if len(vhostsWithoutDefault) > 0 {
		state.AdditionalVHosts = pkg.FromListString(vhostsWithoutDefault)
	} else {
		state.AdditionalVHosts = types.ListNull(types.StringType)
	}

	// TODO: read ENV

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update resource
func (r *ResourceJava) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceJava) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Java

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "JAVA DELETE", map[string]interface{}{"state": state})

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
func (r *ResourceJava) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
