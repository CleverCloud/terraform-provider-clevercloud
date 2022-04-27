package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ResourceNodejsType struct {
	cc  *client.Client
	org string
}

// Create a new resource
func (r ResourceNodejsType) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	app := NodeJS{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &app)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GET variants

	// GET plans

	createAppReq := tmp.CreateAppRequest{
		Name: app.Name.Value,
	}

	res := tmp.CreateApp(ctx, r.cc, r.org, createAppReq)
	if res.HasError() {
		resp.Diagnostics.AddError("failed to create app", res.Error().Error())
		return
	}

	// TODO set fields
	tflog.Info(ctx, "create response", map[string]interface{}{"plan": res.Payload()})

	resp.Diagnostics.Append(resp.State.Set(ctx, app)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information
func (r ResourceNodejsType) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {

}

// Update resource
func (r ResourceNodejsType) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// TODO
}

// Delete resource
func (r ResourceNodejsType) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {

	//resp.State.RemoveResource(ctx)
}

// Import resource
func (r ResourceNodejsType) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
