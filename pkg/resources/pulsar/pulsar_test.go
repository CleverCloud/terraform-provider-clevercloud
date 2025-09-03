package pulsar_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

func TestAccPulsar_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-pulsar")
	fullName := fmt.Sprintf("clevercloud_pulsar.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	pulsarBlock := helper.NewRessource(
		"clevercloud_pulsar",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetAddon(context.Background(), cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect resource '%s' to be deleted", resource.Primary.ID)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(pulsarBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^pulsar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("binary_url"), knownvalue.StringRegexp(regexp.MustCompile(`^pulsar\+ssl:\/\/.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("http_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https:\/\/.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tenant"), knownvalue.StringExact(tests.ORGANISATION)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("namespace"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("token"), knownvalue.NotNull()),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(pulsarBlock.SetOneValue("retention_period", 120).SetOneValue("retention_size", 1024)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^pulsar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("binary_url"), knownvalue.StringRegexp(regexp.MustCompile(`^pulsar\+ssl:\/\/.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("http_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https:\/\/.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tenant"), knownvalue.StringExact(tests.ORGANISATION)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("namespace"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("token"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("retention_period"), knownvalue.Int64Exact(120)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("retention_size"), knownvalue.Int64Exact(1024)),
			},
		}},
	})
}
