package ruby_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// State written by provider 3.0.0 (schema version 1), still holding the
// static_webroot and integrations.redirectionio.backend_port attributes the
// platform never read and that are gone from the schema.
const rubyStateV1WithDroppedAttributes = `{
  "id": "app_4c5d6e7f-8a9b-4c0d-9e1f-2a3b4c5d6e7f",
  "name": "shop",
  "region": "par",
  "vhosts": [{"fqdn": "shop.example.com", "path_begin": "/"}],
  "ruby_version": "3.3",
  "static_files_path": "public",
  "static_url_prefix": "/assets",
  "static_webroot": "public",
  "integrations": {
    "redirectionio": {"project_key": "rio-key", "instance_name": "shop", "backend_port": 9000}
  }
}`

// State written by provider <= 1.x (schema version 0) with static_webroot set.
const rubyStateV0WithStaticWebroot = `{
  "id": "app_4c5d6e7f-8a9b-4c0d-9e1f-2a3b4c5d6e7f",
  "name": "shop",
  "region": "par",
  "vhosts": ["shop.example.com/"],
  "static_files_path": "public",
  "static_webroot": "public"
}`

func TestRuby_UpgradeStateV1_dropsRemovedAttributes(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_ruby", 1, rubyStateV1WithDroppedAttributes)

	if got := tests.StateString(t, state, "static_files_path"); got != "public" {
		t.Errorf("static_files_path = %q, want public", got)
	}

	for name, want := range map[string]string{"project_key": "rio-key", "instance_name": "shop"} {
		path := tftypes.NewAttributePath().
			WithAttributeName("integrations").
			WithAttributeName("redirectionio").
			WithAttributeName(name)
		v, _, err := tftypes.WalkAttributePath(state, path)
		if err != nil {
			t.Fatalf("integrations.redirectionio.%s: %s", name, err)
		}
		var got string
		if err := v.(tftypes.Value).As(&got); err != nil || got != want {
			t.Errorf("integrations.redirectionio.%s = %q (%v), want %q", name, got, err, want)
		}
	}
}

func TestRuby_UpgradeStateV0_dropsStaticWebroot(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_ruby", 0, rubyStateV0WithStaticWebroot)

	if got := tests.StateString(t, state, "static_files_path"); got != "public" {
		t.Errorf("static_files_path = %q, want public", got)
	}
}
