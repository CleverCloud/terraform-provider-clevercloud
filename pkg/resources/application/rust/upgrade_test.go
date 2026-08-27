package rust_test

import (
	"maps"
	"slices"
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// State written by provider <= 1.1.0 (schema version 0): vhosts is a set of
// "fqdn/path" strings, none of the attributes added since exist.
const rustStateV0 = `{
  "id": "app_5e7d2c1a-8f3b-4c6e-9a1d-2b3c4d5e6f70",
  "name": "enhancer",
  "description": "access logs enhancer",
  "min_instance_count": 1,
  "max_instance_count": 2,
  "smallest_flavor": "XS",
  "biggest_flavor": "M",
  "build_flavor": null,
  "region": "par",
  "sticky_sessions": false,
  "redirect_https": true,
  "vhosts": ["console.example.com/", "api.example.com/v1"],
  "app_folder": null,
  "deploy_url": "https://push-n3-par-clevercloud-customers.services.clever-cloud.com/app_5e7d2c1a-8f3b-4c6e-9a1d-2b3c4d5e6f70.git",
  "environment": {"RUST_LOG": "info"},
  "dependencies": [],
  "features": ["tls"],
  "deployment": null,
  "hooks": null
}`

func TestRust_UpgradeStateV0(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_rust", 0, rustStateV0)

	wantVHosts := map[string]string{"console.example.com": "/", "api.example.com": "/v1"}
	if got := tests.StateVHosts(t, state); !maps.Equal(got, wantVHosts) {
		t.Errorf("vhosts = %v, want %v", got, wantVHosts)
	}
	if got := tests.StateString(t, state, "name"); got != "enhancer" {
		t.Errorf("name = %q, want enhancer", got)
	}
	if got := tests.StateStrings(t, state, "features"); !slices.Equal(got, []string{"tls"}) {
		t.Errorf("features = %v, want [tls]", got)
	}
	if !tests.StateAttr(t, state, "exposed_environment").IsNull() {
		t.Error("exposed_environment should be null after upgrade")
	}
}

func TestRust_UpgradeStateV0_nullVHosts(t *testing.T) {
	raw := `{"id": "app_x", "name": "enhancer", "region": "par", "vhosts": null, "features": null}`

	state := tests.UpgradeResourceState(t, "clevercloud_rust", 0, raw)

	if got := tests.StateVHosts(t, state); got != nil {
		t.Errorf("vhosts = %v, want null", got)
	}
}
