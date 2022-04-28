package provider

import (
	"context"

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
	app := NodeJS{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &app)...)
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
		Name:            app.Name.Value,
		Deploy:          "git",
		Description:     app.Description.Value,
		InstanceType:    "node",
		InstanceVariant: variantID,
		InstanceVersion: version,
		MinFlavor:       app.SmallestFlavor.Value,
		MaxFlavor:       app.BiggestFlavor.Value,
		MinInstances:    app.MaxInstanceCount.Value,
		MaxInstances:    app.MaxInstanceCount.Value,
		Zone:            app.Region.Value,
	}

	res := tmp.CreateApp(ctx, r.cc, r.org, createAppReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create app", res.Error().Error())
		return
	}

	appRes := res.Payload()
	// TODO set fields
	tflog.Info(ctx, "create response", map[string]interface{}{"plan": appRes})
	app.ID = fromStr(appRes.ID)
	app.DeployURL = fromStr(appRes.DeployURL)
	app.VHost = fromStr(appRes.Vhosts[0].Fqdn)

	resp.Diagnostics.Append(resp.State.Set(ctx, app)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r ResourceNodeJS) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	tflog.Debug(ctx, "NodeJS READ", map[string]interface{}{"request": req})

	var app NodeJS
	diags := req.State.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appRes := tmp.GetApp(ctx, r.cc, r.org, app.ID.Value)
	if appRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), types.String{Unknown: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if appRes.HasError() {
		resp.Diagnostics.AddError("failed to get app", appRes.Error().Error())
	}

	appNode := appRes.Payload()
	app.DeployURL = fromStr(appNode.DeployURL)
	app.VHost = fromStr(appNode.Vhosts[0].Fqdn)

	diags = resp.State.Set(ctx, app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
