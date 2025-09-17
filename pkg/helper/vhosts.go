package helper

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ToAPIHosts converts a Set of vhosts to []string, adding trailing "/" (format expected by Clever Cloud API).
// Returns an empty slice if null/unknown to let the API handle the cleverapps default domain.
func ToAPIHosts(ctx context.Context, vhosts types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if vhosts.IsNull() || vhosts.IsUnknown() {
		return []string{}, diags // Empty slice -> API handles cleverapps automatically
	}
	var hs []string
	diags = vhosts.ElementsAs(ctx, &hs, false)
	if diags.HasError() {
		return nil, diags
	}

	// Filter cleverapps.io from user-provided vhosts
	filtered := make([]string, 0, len(hs))
	for _, h := range hs {
		normalized := strings.ToLower(strings.TrimSpace(h))
		// Ignore cleverapps.io because the API handles it automatically
		if normalized != "cleverapps.io" {
			if !strings.HasSuffix(h, "/") {
				h = h + "/"
			}
			filtered = append(filtered, h)
		}
	}

	return filtered, diags
}

// FromAPIHosts converts []string (with "/") to []string without trailing "/", normalized.
func FromAPIHosts(raw []string) []string {
	if len(raw) == 0 {
		return []string{"cleverapps.io"}
	}
	out := make([]string, 0, len(raw))
	hasOnlyCleverapps := true
	
	for _, h := range raw {
		h = strings.TrimSpace(h)
		h = strings.TrimSuffix(h, "/")
		h = strings.ToLower(h)
		if h != "" {
			// Check if this is a cleverapps.io domain (format: app-xxx.cleverapps.io)
			if strings.HasSuffix(h, ".cleverapps.io") {
				// If it's only cleverapps domains, we represent it as the generic "cleverapps.io"
				continue
			} else {
				hasOnlyCleverapps = false
				out = append(out, h)
			}
		}
	}
	
	// If we only found cleverapps.io domains, return the generic one
	if hasOnlyCleverapps {
		return []string{"cleverapps.io"}
	}
	
	if len(out) == 0 {
		return []string{"cleverapps.io"}
	}
	sort.Strings(out)
	return out
}

// SetFromStrings builds a Set<string> from a slice
func SetFromStrings(_ context.Context, xs []string) types.Set {
	elems := make([]attr.Value, 0, len(xs))
	for _, s := range xs {
		elems = append(elems, types.StringValue(s))
	}
	v, _ := types.SetValue(types.StringType, elems)
	return v
}

// VHostsFromAPIHosts converts API format vhosts ([]string) to Terraform List of VHost structs
func VHostsFromAPIHosts(raw []string, diags *diag.Diagnostics) types.List {
	if len(raw) == 0 {
		// Return empty list - the API will handle cleverapps.io automatically
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"vhost":      types.StringType,
				"path_begin": types.StringType,
			},
		})
	}

	hasOnlyCleverapps := true
	vhosts := []attr.Value{}

	for _, h := range raw {
		h = strings.TrimSpace(h)
		h = strings.TrimSuffix(h, "/")
		h = strings.ToLower(h)
		
		if h == "" {
			continue
		}

		// Check if this is a cleverapps.io domain (format: app-xxx.cleverapps.io)
		if strings.HasSuffix(h, ".cleverapps.io") {
			// Skip cleverapps domains - we don't include them in the terraform state
			continue
		} else {
			hasOnlyCleverapps = false
			
			// Parse vhost and path_begin from API format
			vhost := h
			pathBegin := "/"
			
			// If the vhost contains a path, extract it
			if idx := strings.Index(h, "/"); idx > 0 {
				vhost = h[:idx]
				pathBegin = h[idx:]
			}
			
			vhostObj, d := types.ObjectValue(
				map[string]attr.Type{
					"vhost":      types.StringType,
					"path_begin": types.StringType,
				},
				map[string]attr.Value{
					"vhost":      types.StringValue(vhost),
					"path_begin": types.StringValue(pathBegin),
				},
			)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"vhost":      types.StringType,
						"path_begin": types.StringType,
					},
				})
			}
			
			vhosts = append(vhosts, vhostObj)
		}
	}

	// If we only found cleverapps.io domains or no vhosts, return null list
	if hasOnlyCleverapps || len(vhosts) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"vhost":      types.StringType,
				"path_begin": types.StringType,
			},
		})
	}

	list, d := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"vhost":      types.StringType,
			"path_begin": types.StringType,
		},
	}, vhosts)
	diags.Append(d...)
	
	return list
}