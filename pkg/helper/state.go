package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TFStruct interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

func PlanFrom[T any](ctx context.Context, p tfsdk.Plan, diags *diag.Diagnostics) T {
	return *From[T](ctx, p, diags)
}

func StateFrom[T any](ctx context.Context, s tfsdk.State, diags *diag.Diagnostics) T {
	return *From[T](ctx, s, diags)
}

func ConfigFrom[T any](ctx context.Context, c tfsdk.Config, diags *diag.Diagnostics) T {
	return *From[T](ctx, c, diags)
}

func From[T any](ctx context.Context, src TFStruct, diags *diag.Diagnostics) *T {
	if src == nil {
		return nil
	}

	var t T
	diags.Append(src.Get(ctx, &t)...)
	return &t
}

// IdentityIDFrom returns the identity "id", or the state "id" when the identity is null
// Terraform/OpenTofu < 1.12 don't store identities
func IdentityIDFrom(ctx context.Context, identity *tfsdk.ResourceIdentity, state tfsdk.State, diags *diag.Diagnostics) types.String {
	var id types.String
	if identity != nil && !identity.Raw.IsNull() {
		diags.Append(identity.GetAttribute(ctx, path.Root("id"), &id)...)
		return id
	}

	diags.Append(state.GetAttribute(ctx, path.Root("id"), &id)...)
	return id
}
