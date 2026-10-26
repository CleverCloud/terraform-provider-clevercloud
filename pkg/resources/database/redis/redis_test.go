package redis_test

import (
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
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/addonid"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.dev/client"
)

func TestAccRedis_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-redis")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_redis.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	materiakvBlock := helper.NewRessource("clevercloud_redis", rName, helper.SetKeyValues(map[string]any{
		"name":   rName,
		"region": "par",
		"plan":   "m_mono",
	}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
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

// TestAccRedis_importAddonID pins that an add-on ID is refused on import: the API
// would accept it almost everywhere, but a network group member registered with it
// never gets a peer. The real ID is what the error has to point at
func TestAccRedis_importAddonID(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-redis")
	fullName := fmt.Sprintf("clevercloud_redis.%s", rName)
	config := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).Append(helper.NewRessource("clevercloud_redis", rName, helper.SetKeyValues(map[string]any{
		"name":   rName,
		"region": "par",
		"plan":   "s_mono",
	}))).String()
	stateID := func(state *terraform.State) string { return state.RootModule().Resources[fullName].Primary.ID }

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			Config: config,
		}, {
			Config:       config,
			ResourceName: fullName,
			ImportState:  true,
			ImportStateIdFunc: func(state *terraform.State) (string, error) {
				cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))
				return addonid.FromRealID(ctx, cc, tests.ORGANISATION, stateID(state))
			},
			ExpectError: regexp.MustCompile(`(?s)cannot\s+import.*redis_`),
		}, {
			Config:       config,
			ResourceName: fullName,
			ImportState:  true,
			ImportStateIdFunc: func(state *terraform.State) (string, error) {
				return stateID(state), nil
			},
			ImportStateVerify: true,
			// Read does not fill creation_date
			ImportStateVerifyIgnore: []string{"creation_date"},
		}},
	})
}
