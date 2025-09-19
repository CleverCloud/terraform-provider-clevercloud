package plan

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"go.clever-cloud.com/terraform-provider/pkg/helper"
)

type NormalizeVHostsPlanModifier struct{}

func (m NormalizeVHostsPlanModifier) Description(ctx context.Context) string {
	return "Normalizes vhosts by removing schema, paths, trailing slash and converting to lowercase"
}
func (m NormalizeVHostsPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m NormalizeVHostsPlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Config unknown or null -> leave as unknown/null
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		resp.PlanValue = req.ConfigValue
		return
	}

	var cfg []string
	diags := req.ConfigValue.ElementsAs(ctx, &cfg, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	norm := make(map[string]struct{}, len(cfg))
	for _, raw := range cfg {
		h, err := NormalizeHost(raw)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid vhost",
				fmt.Sprintf("%q is not a valid vhost: %v", raw, err),
			)
			continue
		}
		if h != "" {
			norm[h] = struct{}{}
		}
	}

	out := make([]string, 0, len(norm))
	for h := range norm {
		out = append(out, h)
	}
	sort.Strings(out) // stable order to avoid cosmetic diffs

	resp.PlanValue = helper.SetFromStrings(ctx, out)
}

// NormalizeVHosts returns a plan modifier that normalizes vhost values
func NormalizeVHosts() planmodifier.Set {
	return NormalizeVHostsPlanModifier{}
}

// NormalizeHost removes schema, cuts everything after / ? #, removes trailing slash and dots, and converts to lowercase.
func NormalizeHost(in string) (string, error) {
	s := strings.TrimSpace(in)
	if s == "" {
		return "", nil
	}
	// try to parse as URL to extract host if schema is present
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		s = u.Host
	} else {
		s = strings.TrimPrefix(s, "http://")
		s = strings.TrimPrefix(s, "https://")
	}
	// cut at first occurrence of path/query/fragment
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSuffix(s, ".")
	s = strings.TrimSuffix(s, "/")
	s = strings.ToLower(strings.TrimSpace(s))

	// minimal validation
	if s == "" || strings.ContainsAny(s, " /") {
		return "", fmt.Errorf("empty or invalid host after normalization")
	}
	return s, nil
}


