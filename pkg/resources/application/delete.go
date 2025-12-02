package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// Delete centralizes the common Delete logic for all application runtimes
func Delete[T RuntimePlan](ctx context.Context, resource RuntimeResource, state T) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Get runtime pointer to access ID
	runtime := state.GetRuntimePtr()

	// Call API to delete the application
	res := tmp.DeleteApp(ctx, resource.Client(), resource.Organization(), runtime.ID.ValueString())

	// If app is already deleted (NotFound), that's OK
	if res.IsNotFoundError() {
		return diags
	}

	// If any other error occurred, report it
	if res.HasError() {
		diags.AddError("failed to delete app", res.Error().Error())
		return diags
	}

	// Success
	return diags
}
