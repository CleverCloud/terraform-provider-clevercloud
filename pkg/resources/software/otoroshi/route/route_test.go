package route_test

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

func TestAccOtoroshiRoute_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-route")
	otoName := acctest.RandomWithPrefix("tf-test-route-oto")
	appName := acctest.RandomWithPrefix("tf-test-route-app")
	fullName := fmt.Sprintf("clevercloud_otoroshi_route.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	otoroshiBlock := helper.NewRessource(
		"clevercloud_otoroshi",
		otoName,
		helper.SetKeyValues(map[string]any{
			"name":         otoName,
			"region":       "par",
			"networkgroup": true,
		}),
	)

	appBlock := helper.NewRessource(
		"clevercloud_static",
		appName,
		helper.SetKeyValues(map[string]any{
			"name":               appName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"networkgroups": []map[string]string{{
				"networkgroup_id": fmt.Sprintf("${clevercloud_otoroshi.%s.networkgroup_id}", otoName),
			}},
		}),
	)

	routeBlock := helper.NewRessource(
		"clevercloud_otoroshi_route",
		rName,
		helper.SetKeyValues(map[string]any{
			"otoroshi_id":    fmt.Sprintf("${clevercloud_otoroshi.%s.id}", otoName),
			"name":           rName,
			"application_id": fmt.Sprintf("${clevercloud_static.%s.id}", appName),
			"port":           4242,
			"path":           "/api",
		}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: "otoroshi_route_" + rName,
			Config:       providerBlock.Append(otoroshiBlock, appBlock, routeBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^route_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("path"), knownvalue.StringExact("/api")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("strip_path"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("enabled"), knownvalue.Bool(true)),
				// The domain defaults to the cleverapps.io host of the Otoroshi entrypoint
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("domain"), knownvalue.StringRegexp(regexp.MustCompile(`^app-.*\.cleverapps\.io$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://app-.*\.cleverapps\.io/api$`))),
				// The backend is the networkgroup FQDN of the target application
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("backend_hostname"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*\.m\.ng_.*\.cc-ng\.cloud$`))),
			},
		}, {
			ResourceName: "otoroshi_route_" + rName,
			Config:       providerBlock.Append(otoroshiBlock, appBlock, routeBlock.SetOneValue("path", "/edited")).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("path"), knownvalue.StringExact("/edited")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://app-.*\.cleverapps\.io/edited$`))),
			},
		}},
	})
}
