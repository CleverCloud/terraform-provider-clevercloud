package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func PlanFrom[T any](ctx context.Context, p tfsdk.Plan, diags *diag.Diagnostics) T {
	var t T
	diags.Append(p.Get(ctx, &t)...)
	return t
}

func StateFrom[T any](ctx context.Context, s tfsdk.State, diags *diag.Diagnostics) T {
	var t T
	diags.Append(s.Get(ctx, &t)...)
	return t
}

func IdentityFrom[T any](ctx context.Context, s tfsdk.ResourceIdentity, diags *diag.Diagnostics) T {
	var t T
	diags.Append(s.Get(ctx, &t)...)
	return t
}
