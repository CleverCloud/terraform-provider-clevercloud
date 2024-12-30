package cellar_test

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

var TestProtoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

func TestAccCellar_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-cellar-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_cellar.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	cellarBlock := helper.NewRessource(
		"clevercloud_cellar",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: TestProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "cellar_" + rName,
			Config:       providerBlock.Append(cellarBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
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
