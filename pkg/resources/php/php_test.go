package php_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccPHP_basic(t *testing.T) {
	t.Logf("starting....")
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-php-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_php.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	phpBlock := helper.NewRessource(
		"clevercloud_php",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"php_version":        "8",
			"additional_vhosts":  [1]string{"toto-tf5283457829345.com"},
		}))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(phpBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^app_.*$`)),
				resource.TestMatchResourceAttr(fullName, "deploy_url", regexp.MustCompile(`^git\+ssh.*\.git$`)),
				resource.TestCheckResourceAttr(fullName, "region", "par"),
			),
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				phpBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "min_instance_count", "2"),
				resource.TestCheckResourceAttr(fullName, "max_instance_count", "6"),
			),
		}},
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetApp(ctx, cc, org, resource.Primary.ID)
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
		},
	})
}
