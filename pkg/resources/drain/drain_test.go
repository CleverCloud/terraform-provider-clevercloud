package drain_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

func TestAccDrainDatadog_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
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
			// Provide a test API key for the datadog recipient
			"api_key": "test-datadog-api-key-abc123",
			// endpoint defaults to US1: https://http-intake.logs.datadoghq.com/v1/input
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
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDrainHTTP_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
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
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDrainSyslogTCP_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
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
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDrainSyslogUDP_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
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
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDrainNewRelic_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_newrelic.%s", rNameDrain)

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
		"clevercloud_drain_newrelic",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind":        "LOG",
			"resource_id": fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			"url":         "https://log-api.newrelic.com/log/v1",
			"api_key":     "test-api-key-123",
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
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("url"), knownvalue.StringExact("https://log-api.newrelic.com/log/v1")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDrainElasticsearch_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_elasticsearch.%s", rNameDrain)

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
		"clevercloud_drain_elasticsearch",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind":             "LOG",
			"resource_id":      fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			"url":              "https://elasticsearch.example.com:9200",
			"username":         "test-user",
			"password":         "test-password",
			"index":            "test-logs",
			"tls_verification": "DEFAULT",
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
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("url"), knownvalue.StringExact("https://elasticsearch.example.com:9200")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("username"), knownvalue.StringExact("test-user")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("index"), knownvalue.StringExact("test-logs")),
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("tls_verification"), knownvalue.StringExact("DEFAULT")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDrainOVH_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	rNameApp := acctest.RandomWithPrefix("tf-static")
	rNameDrain := acctest.RandomWithPrefix("tf-drain")

	fullNameApp := fmt.Sprintf("clevercloud_static.%s", rNameApp)
	fullNameDrain := fmt.Sprintf("clevercloud_drain_ovh.%s", rNameDrain)

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
		"clevercloud_drain_ovh",
		rNameDrain,
		helper.SetKeyValues(map[string]any{
			"kind":        "LOG",
			"resource_id": fmt.Sprintf("${clevercloud_static.%s.id}", rNameApp),
			"url":         "https://gra1.logs.ovh.com/YOUR_LOG_STREAM/",
			"token":       "test-ovh-token-123",
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
				statecheck.ExpectKnownValue(fullNameDrain, tfjsonpath.New("url"), knownvalue.StringExact("https://gra1.logs.ovh.com/YOUR_LOG_STREAM/")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}
