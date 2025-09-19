package helper

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
func VHostsFromAPIHosts(ctx context.Context, raw []string, planValue types.Set, diags *diag.Diagnostics) types.Set {
	schema := map[string]attr.Type{"fqdn": types.StringType, "path_begin": types.StringType}

	if len(raw) == 0 {
		if planValue.IsNull() {
			return types.SetNull(types.ObjectType{AttrTypes: schema})
		}
		return types.SetValueMust(types.ObjectType{AttrTypes: schema}, []attr.Value{})
	}

	vhosts := []attr.Value{}

	for _, h := range raw {
		h = strings.TrimSpace(h)
		hostAndPath := strings.SplitN(h, "/", 2)

		if hostAndPath[0] == "" {
			continue
		}

		vhost := hostAndPath[0]
		pathBegin := "/"
		if len(hostAndPath) > 1 && hostAndPath[1] != "" {
			pathBegin = "/" + hostAndPath[1]
		}

		vhostO := types.ObjectValueMust(schema, map[string]attr.Value{
			"fqdn":       types.StringValue(vhost),
			"path_begin": types.StringValue(pathBegin),
		})
		vhosts = append(vhosts, vhostO)
	}

	list, d := types.SetValue(types.ObjectType{AttrTypes: schema}, vhosts)
	diags.Append(d...)

	return list
}
