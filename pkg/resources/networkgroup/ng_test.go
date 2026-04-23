package networkgroup_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

func TestAccNG_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-ng")
	fullName := fmt.Sprintf("clevercloud_networkgroup.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	addonBlock := helper.NewRessource(
		"clevercloud_networkgroup",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":        rName,
			"description": "par",
			"tags":        []string{"tag1", "tag2"},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^ng_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("description"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("tag1"),
					knownvalue.StringExact("tag2"),
				})),
			},
		}},
	})
}

// TestAccNG_withPeers reproduces issue #337
func TestAccNG_withPeers(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ngName := acctest.RandomWithPrefix("tf-test-ng")
	appName := acctest.RandomWithPrefix("tf-test-docker")
	ngFullName := fmt.Sprintf("clevercloud_networkgroup.%s", ngName)
	appFullName := fmt.Sprintf("clevercloud_docker.%s", appName)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	ngBlock := helper.NewRessource(
		"clevercloud_networkgroup",
		ngName,
		helper.SetKeyValues(map[string]any{
			"name":        ngName,
			"description": "Test networkgroup with peers for issue #337",
			// Intentionally omit "tags" to test that it's optional
		}),
	)
	appBlock := helper.NewRessource(
		"clevercloud_docker",
		appName,
		helper.SetKeyValues(map[string]any{
			"name":               appName,
			"region":             "par",
			"min_instance_count": 0, // Don't start instances
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"networkgroups": []map[string]string{{
				"networkgroup_id": fmt.Sprintf("${clevercloud_networkgroup.%s.id}", ngName),
				"fqdn":            "myapp",
			}},
		}),
	)

	config := providerBlock.Append(ngBlock, appBlock).String()
	t.Logf("Config:\n%s", config)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: ngName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(ngFullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^ng_.*`))),
				statecheck.ExpectKnownValue(ngFullName, tfjsonpath.New("name"), knownvalue.StringExact(ngName)),
				statecheck.ExpectKnownValue(appFullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*`))),
			},
		}, {
			ResourceName:       ngName,
			Config:             config,
			PlanOnly:           true,
			ExpectNonEmptyPlan: false, // We don't expect any changes, just a successful refresh
		}},
	})
}
