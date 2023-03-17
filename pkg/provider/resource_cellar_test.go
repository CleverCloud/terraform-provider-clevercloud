package provider

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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

//go:embed resource_cellar_test_block.tf
var cellarBlock string

func TestAccCellar_basic(t *testing.T) {
	ctx := context.Background()
	cName := fmt.Sprintf("tf-test-cellar-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_cellar.%s", cName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "cellar_" + cName,
			Config:       fmt.Sprintf(providerBlock, org) + fmt.Sprintf(cellarBlock, cName, cName),
			Check: resource.ComposeTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^cellar_.*`)),
				resource.TestMatchResourceAttr(fullName, "host", regexp.MustCompile(`^.*\.services.clever-cloud.com$`)),
				resource.TestMatchResourceAttr(fullName, "key_id", regexp.MustCompile(`^[A-Z0-9]{20}$`)),
				resource.TestMatchResourceAttr(fullName, "key_secret", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
			),
		}},
		CheckDestroy: func(state *terraform.State) error {
			for resourceName, resourceState := range state.RootModule().Resources {
				res := tmp.GetAddon(ctx, cc, org, resourceState.Primary.ID)
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