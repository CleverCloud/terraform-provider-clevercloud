package impl

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/registry"
)

// Resources - Defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return pkg.Map(registry.Resources, func(fn func() resource.Resource) func() resource.Resource {
		return WrapResource(fn)
	})
}

func WrapResource(fn func() resource.Resource) func() resource.Resource {
	return func() resource.Resource {
		r := fn()

		var meta resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "clevercloud"}, &meta)

		return &WrappedResource{resourceName: meta.TypeName, in: r}
	}
}

type WrappedResource struct {
	resourceName string
	in           resource.Resource
}

// Metadata returns the resource type name.
func (w *WrappedResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	w.in.Metadata(ctx, req, resp)
}

// Schema defines the schema for the resource.
func (w *WrappedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	w.in.Schema(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		ev := sentry.NewEvent()
		ev.Message = fmt.Sprintf("%s: Error in resource Schema", w.resourceName)
		ev.Fingerprint = []string{ev.Message}
		ev.Extra = map[string]any{
			"resource": w.resourceName,
		}
		sentry.CaptureEvent(ev)
	}
}

// Create creates a new resource.
func (w *WrappedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	w.in.Create(ctx, req, resp)
	w.handleDiags("Error in resource Create", resp.Diagnostics, map[string]tftypes.Value{
		"plan":   req.Plan.Raw,
		"config": req.Config.Raw,
	})
}

// Read reads the resource.
func (w *WrappedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	w.in.Read(ctx, req, resp)
	w.handleDiags("Error in resource Read", resp.Diagnostics, map[string]tftypes.Value{
		"state": req.State.Raw,
	})
}

// Update updates the resource.
func (w *WrappedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	w.in.Update(ctx, req, resp)
	w.handleDiags("Error in resource Update", resp.Diagnostics, map[string]tftypes.Value{
		"plan":   req.Plan.Raw,
		"config": req.Config.Raw,
		"state":  req.State.Raw,
	})
}

// Delete deletes the resource.
func (w *WrappedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	w.in.Delete(ctx, req, resp)
	w.handleDiags("Error in resource Delete", resp.Diagnostics, map[string]tftypes.Value{
		"state": req.State.Raw,
	})
}

// ImportState handles resource import if the underlying resource implements it.
func (w *WrappedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if importer, ok := w.in.(resource.ResourceWithImportState); ok {
		importer.ImportState(ctx, req, resp)
		w.handleDiags("Error in resource ImportState", resp.Diagnostics, map[string]tftypes.Value{
			//"id":    req.ID, // TODO
			"state": resp.State.Raw,
		})
	}
}

// Configure passes the provider-level configuration data to the resource.
func (w *WrappedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configurer, ok := w.in.(resource.ResourceWithConfigure)
	if !ok {
		return
	}

	configurer.Configure(ctx, req, resp)
	w.handleDiags("Error in resource Configure", resp.Diagnostics, map[string]tftypes.Value{})
}

// UpgradeState upgrades a resource's persisted state if the underlying resource implements it.
func (w *WrappedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if upgrader, ok := w.in.(resource.ResourceWithUpgradeState); ok {
		return upgrader.UpgradeState(ctx)
	}
	return map[int64]resource.StateUpgrader{}
}

func (w *WrappedResource) handleDiags(issueName string, diags diag.Diagnostics, debug map[string]tftypes.Value) {
	if !diags.HasError() {
		return
	}

	ev := sentry.NewEvent()
	ev.Message = fmt.Sprintf("%s: %s", w.resourceName, issueName)
	ev.Fingerprint = []string{ev.Message}
	ev.Extra["resource"] = w.resourceName
	ev.Exception = pkg.Map(diags, func(d diag.Diagnostic) sentry.Exception {
		return sentry.Exception{
			Type:  d.Summary(),
			Value: d.Detail(),
		}
	})

	for k, v := range debug {
		ev.Contexts[k] = helper.SerializeValue(v)
	}

	sentry.CaptureEvent(ev)
}
