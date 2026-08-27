package registry_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.com/terraform-provider/pkg/registry"
)

// Resources whose first published schema version is above 0: no state can exist at
// the versions below, so no upgrader is expected for them.
var firstSchemaVersion = map[string]int64{
	"clevercloud_linux":   1,
	"clevercloud_haskell": 1,
}

// Every resource with a schema version above its first published one must ship a
// state upgrader for each prior version. Without it, Terraform refuses to load any
// state holding an older instance and blocks every operation (#265, #433).
func TestResources_stateUpgradersCoverEveryPriorVersion(t *testing.T) {
	ctx := t.Context()
	seen := map[string]bool{}

	for _, newResource := range registry.Resources {
		r := newResource()

		meta := resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "clevercloud"}, &meta)
		seen[meta.TypeName] = true

		sch := resource.SchemaResponse{}
		r.Schema(ctx, resource.SchemaRequest{}, &sch)

		first := firstSchemaVersion[meta.TypeName]
		if sch.Schema.Version < first {
			t.Errorf("%s: firstSchemaVersion says %d but schema version is %d", meta.TypeName, first, sch.Schema.Version)
			continue
		}

		upgraders := map[int64]resource.StateUpgrader{}
		if withUpgrade, ok := r.(resource.ResourceWithUpgradeState); ok {
			upgraders = withUpgrade.UpgradeState(ctx)
		}

		for version := first; version < sch.Schema.Version; version++ {
			if _, ok := upgraders[version]; !ok {
				t.Errorf("%s: schema version %d but no state upgrader for version %d", meta.TypeName, sch.Schema.Version, version)
			}
		}
		for version := range upgraders {
			if version >= sch.Schema.Version {
				t.Errorf("%s: state upgrader for version %d is never called, schema version is %d", meta.TypeName, version, sch.Schema.Version)
			}
		}
	}

	for typeName := range firstSchemaVersion {
		if !seen[typeName] {
			t.Errorf("firstSchemaVersion lists %s, which is not a registered resource", typeName)
		}
	}
}
