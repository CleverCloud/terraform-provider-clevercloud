package dotnet_test

import (
	"maps"
	"testing"

	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// State written by provider 1.3.0 (schema version 0): the runtime was created after
// the vhosts refactor, so vhosts are already {fqdn, path_begin} objects. Only the
// attributes added later are missing.
const dotnetStateV0 = `{
  "id": "app_2c3d4e5f-6a7b-4c8d-9e0f-1a2b3c4d5e6f",
  "name": "api",
  "description": "dotnet api",
  "min_instance_count": 1,
  "max_instance_count": 3,
  "smallest_flavor": "S",
  "biggest_flavor": "L",
  "build_flavor": null,
  "region": "par",
  "sticky_sessions": false,
  "redirect_https": true,
  "vhosts": [{"fqdn": "api.example.com", "path_begin": "/"}, {"fqdn": "admin.example.com", "path_begin": "/v2"}],
  "app_folder": null,
  "deploy_url": "https://push-n3-par-clevercloud-customers.services.clever-cloud.com/app_2c3d4e5f-6a7b-4c8d-9e0f-1a2b3c4d5e6f.git",
  "environment": {"ASPNETCORE_ENVIRONMENT": "Production"},
  "dependencies": [],
  "profile": "Release",
  "proj": "Api",
  "tfm": "net8.0",
  "version": "8.0",
  "deployment": {"repository": "https://github.com/example/api.git", "commit": "0123456789abcdef0123456789abcdef01234567"},
  "hooks": {"pre_build": null, "post_build": "dotnet test", "pre_run": null, "run_succeed": null, "run_failed": null}
}`

// State written by provider 1.6.0 to 1.7.x, still schema version 0 but with the
// networkgroups attribute that arrived in the meantime.
const dotnetStateV0WithNetworkgroups = `{
  "id": "app_2c3d4e5f-6a7b-4c8d-9e0f-1a2b3c4d5e6f",
  "name": "api",
  "region": "par",
  "vhosts": [{"fqdn": "api.example.com", "path_begin": "/"}],
  "networkgroups": [{"networkgroup_id": "ng_3a4b5c6d-7e8f-4a9b-8c0d-1e2f3a4b5c6d", "fqdn": "api.ng.example.com"}],
  "version": "8.0"
}`

func TestDotnet_UpgradeStateV0(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_dotnet", 0, dotnetStateV0)

	// vhosts were already nested objects, they must go through untouched
	wantVHosts := map[string]string{"api.example.com": "/", "admin.example.com": "/v2"}
	if got := tests.StateVHosts(t, state); !maps.Equal(got, wantVHosts) {
		t.Errorf("vhosts = %v, want %v", got, wantVHosts)
	}
	if got := tests.StateString(t, state, "version"); got != "8.0" {
		t.Errorf("version = %q, want 8.0", got)
	}
	if got := tests.StateString(t, state, "proj"); got != "Api" {
		t.Errorf("proj = %q, want Api", got)
	}
	if got := tests.StateNestedString(t, state, "deployment", "commit"); got != "0123456789abcdef0123456789abcdef01234567" {
		t.Errorf("deployment.commit = %q", got)
	}
	if got := tests.StateNestedString(t, state, "hooks", "post_build"); got != "dotnet test" {
		t.Errorf("hooks.post_build = %q", got)
	}
	if !tests.StateAttr(t, state, "exposed_environment").IsNull() {
		t.Error("exposed_environment should be null after upgrade")
	}
	if !tests.StateAttr(t, state, "networkgroups").IsNull() {
		t.Error("networkgroups should be null when absent from the old state")
	}
}

func TestDotnet_UpgradeStateV0_keepsNetworkgroups(t *testing.T) {
	state := tests.UpgradeResourceState(t, "clevercloud_dotnet", 0, dotnetStateV0WithNetworkgroups)

	got := tests.StateObjects(t, state, "networkgroups")
	want := []map[string]string{{"networkgroup_id": "ng_3a4b5c6d-7e8f-4a9b-8c0d-1e2f3a4b5c6d", "fqdn": "api.ng.example.com"}}
	if len(got) != 1 || !maps.Equal(got[0], want[0]) {
		t.Errorf("networkgroups = %v, want %v", got, want)
	}
}
