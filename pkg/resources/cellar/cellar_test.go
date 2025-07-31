package cellar_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"
	"time"

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

func TestAccCellar_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-cellar-%d", time.Now().UnixMilli())
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_cellar.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	cellarBlock := helper.NewRessource(
		"clevercloud_cellar",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "cellar_" + rName,
			Config:       providerBlock.Append(cellarBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^cellar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Z0-9]{20}$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
			},
		}, {
			ResourceName: "cellar_" + rName,
			Config:       providerBlock.Append(cellarBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^cellar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Z0-9]{20}$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
			},
		}},
		CheckDestroy: func(state *terraform.State) error {
			for resourceName, resourceState := range state.RootModule().Resources {
				res := tmp.GetAddon(ctx, cc, tests.ORGANISATION, resourceState.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect cellar resource '%s' to be deleted: %+v", resourceName, res.Payload())
			}
			return nil
		},
	})
}
