package docker_test

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



func TestAccDocker_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-docker-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_docker.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	dockerBlock := helper.NewRessource(
		"clevercloud_docker",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
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
			Destroy:      false,
			ResourceName: rName,
			Config:       providerBlock.Append(dockerBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^app_.*$`)),
				resource.TestMatchResourceAttr(fullName, "deploy_url", regexp.MustCompile(`^git\+ssh.*\.git$`)),
				resource.TestCheckResourceAttr(fullName, "region", "par"),
			),
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				dockerBlock.SetOneValue("biggest_flavor", "XS"),
			).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "biggest_flavor", "XS"),
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
