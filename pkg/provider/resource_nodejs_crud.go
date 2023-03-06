package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

type ResourceNodeJS struct {
	cc  *client.Client
	org string
}

func init() {
	AddResource(NewResourceNodeJS)
}

func NewResourceNodeJS() resource.Resource {
	return &ResourceNodeJS{}
}

func (r *ResourceNodeJS) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_nodejs"
}

// Weird behaviour, but TF can ask for a Resource without having configured a Provider (maybe for Meta and Schema)
// So we need to handle the case there is no ProviderData
func (r *ResourceNodeJS) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "ResourceNodeJS.Configure()")

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
func (r *ResourceNodeJS) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		Name:            app.Name.ValueString(),
		Deploy:          "git",
		Description:     app.Description.ValueString(),
		InstanceType:    "node",
		InstanceVariant: variantID,
		InstanceVersion: version,
		MinFlavor:       app.SmallestFlavor.ValueString(),
		MaxFlavor:       app.BiggestFlavor.ValueString(),
		MinInstances:    app.MaxInstanceCount.ValueInt64(),
		MaxInstances:    app.MaxInstanceCount.ValueInt64(),
		Zone:            app.Region.ValueString(),
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
func (r *ResourceNodeJS) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "NodeJS READ", map[string]interface{}{"request": req})

	var app NodeJS
	diags := req.State.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appRes := tmp.GetApp(ctx, r.cc, r.org, app.ID.ValueString())
	if appRes.IsNotFoundError() {
		diags = resp.State.SetAttribute(ctx, path.Root("id"), types.StringUnknown)
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
func (r *ResourceNodeJS) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// TODO
}

// Delete resource
func (r *ResourceNodeJS) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var app NodeJS

	diags := req.State.Get(ctx, &app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "NodeJS DELETE", map[string]interface{}{"app": app})

	res := tmp.DeleteApp(ctx, r.cc, r.org, app.ID.ValueString())
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
func (r *ResourceNodeJS) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// and call Read() to fill fields
	attr := path.Root("id")
	resource.ImportStatePassthroughID(ctx, attr, req, resp)
}
