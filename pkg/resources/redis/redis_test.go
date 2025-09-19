package redis_test

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

func TestAccRedis_basic(t *testing.T) {
	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-redis")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_redis.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	materiakvBlock := helper.NewRessource("clevercloud_redis", rName, helper.SetKeyValues(map[string]any{
		"name":   rName,
		"region": "par",
		"plan":   "m_mono",
	}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetAddon(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect resource '%s' to be deleted: %+v", resource.Primary.ID, res.Payload())
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^redis_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^redis_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}},
	})
}
