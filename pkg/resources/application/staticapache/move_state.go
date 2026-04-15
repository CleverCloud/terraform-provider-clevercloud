package staticapache

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

// MoveState enables a zero-downtime migration from the legacy clevercloud_static
// resource (which, before v1.12.0, actually managed Static-with-Apache apps under
// variant "static-apache") to the new, correctly-named clevercloud_static_apache.
//
// Users opt in by adding a `moved` block in their HCL:
//
//	moved {
//	  from = clevercloud_static.myapp
//	  to   = clevercloud_static_apache.myapp
//	}
//
// The schemas are structurally identical (both embed application.Runtime), so the
// move is a straight state re-tagging — no attribute transformation needed.
func (r *ResourceStaticApache) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			SourceSchema: &schemaStaticApache,
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				// Only handle moves from the legacy clevercloud_static resource.
				// Other sources are ignored so Terraform can try the next mover
				// (or surface its own error if none matches).
				if req.SourceTypeName != "clevercloud_static" {
					return
				}

				source := helper.StateFrom[StaticApache](ctx, *req.SourceState, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.TargetState.Set(ctx, &source)...)
			},
		},
	}
}
