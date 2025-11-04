package defaultloadbalancer_test

import (
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

func TestAccDataSourceDefaultLoadBalancer_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-lb")
	appResourceName := fmt.Sprintf("clevercloud_docker.%s", rName)
	dsName := fmt.Sprintf("data.clevercloud_default_loadbalancer.%s", rName)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Create a simple Docker application to test the datasource
	dockerBlock := helper.NewRessource(
		"clevercloud_docker",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}))

	// Create the datasource block that references the application
	lbDataBlock := helper.NewDataRessource(
		"clevercloud_default_loadbalancer",
		rName,
		helper.SetKeyValues(map[string]any{
			"application_id": fmt.Sprintf("${%s.id}", appResourceName),
		}))

	config := providerBlock.Append(dockerBlock, lbDataBlock).String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			Config: config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(appResourceName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),

				statecheck.ExpectKnownValue(dsName, tfjsonpath.New("name"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(dsName, tfjsonpath.New("cname"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.par\.clever-cloud\.com\.$`))),
				statecheck.ExpectKnownValue(dsName, tfjsonpath.New("servers"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(dsName, tfjsonpath.New("servers"), knownvalue.ListPartial(map[int]knownvalue.Check{
					0: knownvalue.StringRegexp(ipv4_regex),
				})),
			},
		}},
	})
}

var ipv4_regex = regexp.MustCompile(`^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`)
