package frankenphp_test

import (
	"maps"
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// State written by provider 1.1.0 (schema version 0): vhosts is a set of
// "fqdn/path" strings, none of the attributes added since exist.
const frankenphpStateV0 = `{
  "id": "app_9b8a7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d",
  "name": "website",
  "description": null,
  "min_instance_count": 1,
  "max_instance_count": 1,
  "smallest_flavor": "S",
  "biggest_flavor": "S",
  "build_flavor": "M",
  "region": "par",
  "sticky_sessions": true,
  "redirect_https": true,
  "vhosts": ["www.example.com/", "example.com/"],
  "app_folder": "site",
  "deploy_url": "https://push-n3-par-clevercloud-customers.services.clever-cloud.com/app_9b8a7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d.git",
  "environment": {},
  "dependencies": ["addon_1f2e3d4c-5b6a-4978-8f9e-0a1b2c3d4e5f"],
  "dev_dependencies": true,
  "deployment": {"repository": "https://github.com/example/site.git", "commit": null},
  "hooks": {"pre_build": "composer install", "post_build": null, "pre_run": null, "run_succeed": null, "run_failed": null}
}`

func TestFrankenPHP_UpgradeStateV0(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_frankenphp", 0, frankenphpStateV0)

	wantVHosts := map[string]string{"www.example.com": "/", "example.com": "/"}
	if got := tests.StateVHosts(t, state); !maps.Equal(got, wantVHosts) {
		t.Errorf("vhosts = %v, want %v", got, wantVHosts)
	}
	if got := tests.StateString(t, state, "app_folder"); got != "site" {
		t.Errorf("app_folder = %q, want site", got)
	}

	var devDeps bool
	if err := tests.StateAttr(t, state, "dev_dependencies").As(&devDeps); err != nil || !devDeps {
		t.Errorf("dev_dependencies = %v (%v), want true", devDeps, err)
	}
	if !tests.StateAttr(t, state, "exposed_environment").IsNull() {
		t.Error("exposed_environment should be null after upgrade")
	}
}

func TestFrankenPHP_UpgradeStateV0_keepsBlocks(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_frankenphp", 0, frankenphpStateV0)

	if got := tests.StateNestedString(t, state, "deployment", "repository"); got != "https://github.com/example/site.git" {
		t.Errorf("deployment.repository = %q", got)
	}
	if got := tests.StateNestedString(t, state, "hooks", "pre_build"); got != "composer install" {
		t.Errorf("hooks.pre_build = %q", got)
	}
}

func TestFrankenPHP_UpgradeStateV0_nullVHosts(t *testing.T) {
	raw := `{"id": "app_x", "name": "website", "region": "par", "vhosts": null}`

	state := tests.UpgradeResourceState(t, "clevercloud_frankenphp", 0, raw)

	if got := tests.StateVHosts(t, state); got != nil {
		t.Errorf("vhosts = %v, want null", got)
	}
}
