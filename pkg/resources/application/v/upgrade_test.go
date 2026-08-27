package v_test

import (
	"maps"
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// State written by provider 1.3.0 (schema version 0): the runtime was created after
// the vhosts refactor, so vhosts are already {fqdn, path_begin} objects. Only the
// attributes added later are missing.
const vStateV0 = `{
  "id": "app_7a8b9c0d-1e2f-4a3b-8c4d-5e6f7a8b9c0d",
  "name": "vweb",
  "description": null,
  "min_instance_count": 1,
  "max_instance_count": 1,
  "smallest_flavor": "XS",
  "biggest_flavor": "XS",
  "build_flavor": null,
  "region": "par",
  "sticky_sessions": false,
  "redirect_https": false,
  "vhosts": [{"fqdn": "v.example.com", "path_begin": "/"}],
  "app_folder": "web",
  "deploy_url": "https://push-n3-par-clevercloud-customers.services.clever-cloud.com/app_7a8b9c0d-1e2f-4a3b-8c4d-5e6f7a8b9c0d.git",
  "environment": {},
  "dependencies": [],
  "binary": "server",
  "development_build": true,
  "deployment": {"repository": "https://github.com/example/vweb.git", "commit": null},
  "hooks": {"pre_build": null, "post_build": null, "pre_run": "v -version", "run_succeed": null, "run_failed": null}
}`

// State written by provider 1.6.0 to 1.7.x, still schema version 0 but with the
// networkgroups attribute that arrived in the meantime.
const vStateV0WithNetworkgroups = `{
  "id": "app_7a8b9c0d-1e2f-4a3b-8c4d-5e6f7a8b9c0d",
  "name": "vweb",
  "region": "par",
  "vhosts": [{"fqdn": "v.example.com", "path_begin": "/"}],
  "networkgroups": [{"networkgroup_id": "ng_3a4b5c6d-7e8f-4a9b-8c0d-1e2f3a4b5c6d", "fqdn": "vweb.ng.example.com"}],
  "binary": "server"
}`

func TestV_UpgradeStateV0(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_v", 0, vStateV0)

	// vhosts were already nested objects, they must go through untouched
	wantVHosts := map[string]string{"v.example.com": "/"}
	if got := tests.StateVHosts(t, state); !maps.Equal(got, wantVHosts) {
		t.Errorf("vhosts = %v, want %v", got, wantVHosts)
	}
	if got := tests.StateString(t, state, "binary"); got != "server" {
		t.Errorf("binary = %q, want server", got)
	}
	var devBuild bool
	if err := tests.StateAttr(t, state, "development_build").As(&devBuild); err != nil || !devBuild {
		t.Errorf("development_build = %v (%v), want true", devBuild, err)
	}
	if got := tests.StateNestedString(t, state, "deployment", "repository"); got != "https://github.com/example/vweb.git" {
		t.Errorf("deployment.repository = %q", got)
	}
	if got := tests.StateNestedString(t, state, "hooks", "pre_run"); got != "v -version" {
		t.Errorf("hooks.pre_run = %q", got)
	}
	if !tests.StateAttr(t, state, "exposed_environment").IsNull() {
		t.Error("exposed_environment should be null after upgrade")
	}
}

func TestV_UpgradeStateV0_keepsNetworkgroups(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_v", 0, vStateV0WithNetworkgroups)

	got := tests.StateObjects(t, state, "networkgroups")
	want := []map[string]string{{"networkgroup_id": "ng_3a4b5c6d-7e8f-4a9b-8c0d-1e2f3a4b5c6d", "fqdn": "vweb.ng.example.com"}}
	if len(got) != 1 || !maps.Equal(got[0], want[0]) {
		t.Errorf("networkgroups = %v, want %v", got, want)
	}
}
