package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
