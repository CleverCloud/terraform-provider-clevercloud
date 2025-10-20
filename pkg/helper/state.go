package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type TFStruct interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

func PlanFrom[T any](ctx context.Context, p tfsdk.Plan, diags *diag.Diagnostics) T {
	return From[T](ctx, p, diags)
}

func StateFrom[T any](ctx context.Context, s tfsdk.State, diags *diag.Diagnostics) T {
	return From[T](ctx, s, diags)
}

func IdentityFrom[T any](ctx context.Context, s tfsdk.ResourceIdentity, diags *diag.Diagnostics) T {
	return From[T](ctx, s, diags)
}

func From[T any](ctx context.Context, src TFStruct, diags *diag.Diagnostics) T {
	var t T
	diags.Append(src.Get(ctx, &t)...)
	return t
}
