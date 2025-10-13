package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.clever-cloud.com/terraform-provider/pkg/sweepers"
)

// TestSweepers verifies that sweepers are properly registered
// This test doesn't actually run the sweepers, just checks they exist
func TestMain(m *testing.M) {
	resource.AddTestSweepers("clevercloud_networkgroup", &resource.Sweeper{
		Name: "clevercloud_networkgroup",
		F:    sweepers.SweepNetworkgroups,
	})

	resource.AddTestSweepers("clevercloud_application", &resource.Sweeper{
		Name: "clevercloud_application",
		F:    sweepers.SweepApplications,
		Dependencies: []string{
			"clevercloud_networkgroup", // Apps might be in network groups
		},
	})

	resource.AddTestSweepers("clevercloud_addon", &resource.Sweeper{
		Name: "clevercloud_addon",
		F:    sweepers.SweepAddons,
		Dependencies: []string{
			"clevercloud_application", // Apps might be linked to addons
		},
	})

	fmt.Printf("sweepers added, running tests\n")
	resource.TestMain(m)
}
