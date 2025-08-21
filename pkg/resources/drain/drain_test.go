package drain_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccDrainDatadog_basic(t *testing.T) {
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_datadog.%s", rNameDrain)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	staticBlock := helper.NewRessource(
		"clevercloud_static",
		rNameApp,
		helper.SetKeyValues(map[string]any{
			"name":               rNameApp,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}),
	)

	drainBlock := helper.NewRessource(
		"clevercloud_drain_datadog",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind": "LOG",
			// Reference the application ID created above
			"resource_id": fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			// Provide a placeholder URL for the datadog recipient
			"url": "https://example.com/datadog",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rNameDrain,
			Config:       providerBlock.Append(staticBlock, drainBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				// App assertions
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("region"), knownvalue.StringExact("par")),

				// Drain assertions
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("kind"), knownvalue.StringExact("LOG")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("resource_id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
			},
		}},
		CheckDestroy: checkDrainDestroyed,
	})
}

func TestAccDrainHTTP_basic(t *testing.T) {
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_http.%s", rNameDrain)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	staticBlock := helper.NewRessource(
		"clevercloud_static",
		rNameApp,
		helper.SetKeyValues(map[string]any{
			"name":               rNameApp,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}),
	)

	drainBlock := helper.NewRessource(
		"clevercloud_drain_http",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind":        "LOG",
			"resource_id": fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			"url":         "https://httpbin.org/post", // Public HTTP endpoint for testing
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rNameDrain,
			Config:       providerBlock.Append(staticBlock, drainBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				// App assertions
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("region"), knownvalue.StringExact("par")),

				// Drain assertions
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("kind"), knownvalue.StringExact("LOG")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("resource_id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("url"), knownvalue.StringExact("https://httpbin.org/post")),
			},
		}},
		CheckDestroy: checkDrainDestroyed,
	})
}

func TestAccDrainSyslogTCP_basic(t *testing.T) {
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_syslog_tcp.%s", rNameDrain)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	staticBlock := helper.NewRessource(
		"clevercloud_static",
		rNameApp,
		helper.SetKeyValues(map[string]any{
			"name":               rNameApp,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}),
	)

	drainBlock := helper.NewRessource(
		"clevercloud_drain_syslog_tcp",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind":        "LOG",
			"resource_id": fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			"url":         "tcp://test.example.com:514",
			"token":       "test-token",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rNameDrain,
			Config:       providerBlock.Append(staticBlock, drainBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				// App assertions
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("region"), knownvalue.StringExact("par")),

				// Drain assertions
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("kind"), knownvalue.StringExact("LOG")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("resource_id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("url"), knownvalue.StringExact("tcp://test.example.com:514")),
			},
		}},
		CheckDestroy: checkDrainDestroyed,
	})
}

func TestAccDrainSyslogUDP_basic(t *testing.T) {
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_syslog_udp.%s", rNameDrain)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	staticBlock := helper.NewRessource(
		"clevercloud_static",
		rNameApp,
		helper.SetKeyValues(map[string]any{
			"name":               rNameApp,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}),
	)

	drainBlock := helper.NewRessource(
		"clevercloud_drain_syslog_udp",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind":        "LOG",
			"resource_id": fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			"url":         "udp://test.example.com:514",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rNameDrain,
			Config:       providerBlock.Append(staticBlock, drainBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				// App assertions
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameApp, tfjsonpath.New("region"), knownvalue.StringExact("par")),

				// Drain assertions
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("kind"), knownvalue.StringExact("LOG")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("resource_id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("url"), knownvalue.StringExact("udp://test.example.com:514")),
			},
		}},
		CheckDestroy: checkDrainDestroyed,
	})
}

// Common function to check drain destruction
func checkDrainDestroyed(state *terraform.State) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig())

	for name, resource := range state.RootModule().Resources {
		fmt.Printf("Checking drain '%s'\n", name)

		res := tmp.GetApp(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
		if res.IsNotFoundError() {
			continue
		}
		if res.HasError() {
			return fmt.Errorf("unexpectd error: %s", res.Error().Error())
		}
		if res.Payload().State == "TO_DELETE" {
			continue
		}

		return fmt.Errorf("expect resource '%s' to be deleted state: '%s'", resource.Primary.ID, res.Payload().State)
	}
	return nil
}
