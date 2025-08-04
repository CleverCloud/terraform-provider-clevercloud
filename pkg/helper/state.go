package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type TFStruct interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

func From[T any](ctx context.Context, src TFStruct, diags *diag.Diagnostics) (t T) {
	diags.Append(src.Get(ctx, &t)...)
	return
}

// Kept for retrocompat
func PlanFrom[T any](ctx context.Context, p tfsdk.Plan, diags *diag.Diagnostics) T {
	return From[T](ctx, p, diags)
}

// Kept for retrocompat
func StateFrom[T any](ctx context.Context, s tfsdk.State, diags *diag.Diagnostics) T {
	return From[T](ctx, s, diags)
}
