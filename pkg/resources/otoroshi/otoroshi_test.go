package otoroshi_test

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

func TestAccOtoroshi_basic(t *testing.T) {
	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-otoroshi")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_otoroshi.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	otoroshiBlock := helper.NewRessource(
		"clevercloud_otoroshi",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "otoroshi_" + rName,
			Config:       providerBlock.Append(otoroshiBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^otoroshi_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-api-otoroshi\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_login"), knownvalue.StringExact("cc-account-admin")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_password"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-ui-otoroshi\.services\.clever-cloud\.com$`))),
			},
		}, {
			ResourceName: "otoroshi_" + rName,
			Config:       providerBlock.Append(otoroshiBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^otoroshi_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-api-otoroshi\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_login"), knownvalue.StringExact("cc-account-admin")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_password"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-ui-otoroshi\.services\.clever-cloud\.com$`))),
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

				return fmt.Errorf("expect otoroshi resource '%s' to be deleted: %+v", resourceName, res.Payload())
			}
			return nil
		},
	})
}
