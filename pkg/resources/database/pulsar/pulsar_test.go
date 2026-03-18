package pulsar_test

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

func TestAccPulsar_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-pulsar")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_pulsar.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	pulsarBlock := helper.NewRessource(
		"clevercloud_pulsar",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
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
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(pulsarBlock.SetOneValue("name", rNameEdited).SetOneValue("retention_period", 120).SetOneValue("retention_size", 1024)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^pulsar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("binary_url"), knownvalue.StringRegexp(regexp.MustCompile(`^pulsar\+ssl:\/\/.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("http_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https:\/\/.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tenant"), knownvalue.StringExact(tests.ORGANISATION)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("namespace"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("retention_period"), knownvalue.Int64Exact(120)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("retention_size"), knownvalue.Int64Exact(1024)),
			},
		}},
	})
}
