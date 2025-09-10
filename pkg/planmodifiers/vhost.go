package planmodifiers

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// VhostNormalizer returns a plan modifier that normalizes vhost values by removing trailing slashes
// This ensures transparent handling for users regardless of whether they include trailing slashes
func VhostNormalizer() planmodifier.Set {
	return vhostNormalizerModifier{}
}

// vhostNormalizerModifier implements planmodifier.Set
type vhostNormalizerModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m vhostNormalizerModifier) Description(_ context.Context) string {
	return "Normalizes vhost values by removing trailing slashes for transparent user experience"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m vhostNormalizerModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes vhost values by removing trailing slashes for transparent user experience"
}

// PlanModifySet implements the plan modification logic.
func (m vhostNormalizerModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing if there is no plan value
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	// Extract the current plan values
	var vhosts []string
	diags := req.PlanValue.ElementsAs(ctx, &vhosts, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normalize vhosts by removing trailing slashes
	normalizedVhosts := make([]string, len(vhosts))
	modified := false
	
	for i, vhost := range vhosts {
		normalized := strings.TrimSuffix(vhost, "/")
		normalizedVhosts[i] = normalized
		if normalized != vhost {
			modified = true
		}
	}

	// Only update the plan if normalization changed values
	if modified {
		tflog.Debug(ctx, "VhostNormalizer: Normalizing vhosts", map[string]any{
			"original":   vhosts,
			"normalized": normalizedVhosts,
		})
		
		normalizedSet, diags := types.SetValueFrom(ctx, types.StringType, normalizedVhosts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		resp.PlanValue = normalizedSet
	}
}