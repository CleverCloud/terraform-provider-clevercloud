package addon_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

func TestAccAddon_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-mp-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_addon.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org).String()
	addonBlock := helper.NewRessource(
		"clevercloud_addon",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":                 rName,
			"region":               "par",
			"plan":                 "clever_solo",
			"third_party_provider": "mailpace",
		})).String()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetAddon(context.Background(), cc, org, resource.Primary.ID)
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
			Config:       providerBlock + addonBlock,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^addon_.*`)),
				//resource.TestMatchResourceAttr(fullName, "password", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				// TODO test env var existance
			),
		}},
	})
}
