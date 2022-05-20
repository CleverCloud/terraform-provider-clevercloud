package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ResourceNodeJS struct {
	cc  *client.Client
	org string
}

// Create a new resource
func (r ResourceNodeJS) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	state := NodeJS{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
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
		if product.Type != "node" || product.Name != "Node" {
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

	createAppReq := tmp.CreateAppRequest{
		Name:            state.Name.Value,
		Deploy:          "git",
		Description:     state.Description.Value,
		InstanceType:    "node",
		InstanceVariant: variantID,
		InstanceVersion: version,
		MinFlavor:       state.SmallestFlavor.Value,
		MaxFlavor:       state.BiggestFlavor.Value,
		MinInstances:    state.MinInstanceCount.Value,
		MaxInstances:    state.MaxInstanceCount.Value,
		Zone:            state.Region.Value,
	}

	res := tmp.CreateApp(ctx, r.cc, r.org, createAppReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create app", res.Error().Error())
		return
	}

	appRes := res.Payload()
	tflog.Info(ctx, "create response", map[string]interface{}{"res": appRes})
	state.ID = fromStr(appRes.ID)
	state.VHost = fromStr(appRes.Vhosts[0].Fqdn)
	state.DeployURL = fromStr(appRes.DeployURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set env vars
	planEnvs, diags := state.GetEnv(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envRes := tmp.UpdateAppEnv(ctx, r.cc, r.org, state.ID.Value, planEnvs)
	if envRes.HasError() {
		// Set empty in state, then if apply again, will retry
		state.Environment = types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}}
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

		resp.Diagnostics.AddWarning("failed to configure env vars", envRes.Error().Error())
		return
	}

	var dependencies []string
	state.Dependencies.ElementsAs(ctx, &dependencies, false)
	for _, dependency := range dependencies {
		depRes := tmp.CreateDependency(ctx, r.cc, r.org, state.ID.Value, dependency)
		if depRes.HasError() {
			resp.Diagnostics.AddError("failed to link app and addon", depRes.Error().Error())
		}
	}
}

// Read resource information
func (r ResourceNodeJS) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	tflog.Debug(ctx, "NodeJS READ", map[string]interface{}{"request": req})

	state := NodeJS{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.read(ctx, &resp.State, &state)...)
}

// app.ID need to be set
func (r ResourceNodeJS) read(ctx context.Context, state *tfsdk.State, app *NodeJS) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// main app
	appRes := tmp.GetApp(ctx, r.cc, r.org, app.ID.Value)
	if appRes.IsNotFoundError() {
		diags = state.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), types.String{Unknown: true})
		diags.Append(diags...)
		if diags.HasError() {
			return diags
		}
	}
	if appRes.HasError() {
		diags.AddError("failed to get app", appRes.Error().Error())
	} else {
		appNode := appRes.Payload()
		app.ID = fromStr(appNode.ID)
		app.Name = fromStr(appNode.Name)
		app.Description = fromStr(appNode.Description)
		app.MinInstanceCount = fromI(appNode.Instance.MinInstances)
		app.MaxInstanceCount = fromI(appNode.Instance.MaxInstances)
		app.SmallestFlavor = fromStr(appNode.Instance.MinFlavor.Name)
		app.BiggestFlavor = fromStr(appNode.Instance.MaxFlavor.Name)
		app.Region = fromStr(appNode.Zone)
		app.VHost = fromStr(appNode.Vhosts[0].Fqdn)
		app.DeployURL = fromStr(appNode.DeployURL)

		diags.Append(state.Set(ctx, app)...)
		fmt.Println("###################### STATE SAVED ############################")
	}

	// app env
	envRes := tmp.ListAppEnv(ctx, r.cc, r.org, app.ID.Value)
	if envRes.HasError() {
		diags.AddError("failed to get app env", envRes.Error().Error())
	} else {
		for _, env := range *envRes.Payload() {
			app.Environment.Elems[env.Name] = fromStr(env.Value)
		}

		diags.Append(state.Set(ctx, app)...)
	}

	// app deps
	/*depRes := tmp.ListDependencies(ctx, r.cc, r.org, app.ID.Value)
	if depRes.HasError() {
		diags.AddError("failed to list app dependencies", depRes.Error().Error())
	} else {
		app.Dependencies = pkg.ReduceList(
			*depRes.Payload(),
			types.List{ElemType: types.StringType},
			func(acc types.List, item tmp.AddonResponse) types.List {
				strVal := fromStr(item.ID)
				acc.Elems = append(acc.Elems, strVal)

				return acc
			},
		)

		diags.Append(state.Set(ctx, app)...)
	}*/

	return diags
}

// Update resource
func (r ResourceNodeJS) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// TODO
}

// Delete resource
func (r ResourceNodeJS) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var app NodeJS

	diags := req.State.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "NodeJS DELETE", map[string]interface{}{"app": app})

	res := tmp.DeleteApp(ctx, r.cc, r.org, app.ID.Value)
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
func (r ResourceNodeJS) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
