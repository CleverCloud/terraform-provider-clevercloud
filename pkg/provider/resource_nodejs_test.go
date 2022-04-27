package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

const nodejsBlock = `resource "%s" "%s" { 
	name = "%s"
	plan = "dev" 
	region = "par"
}
`

func TestAccNodejs_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-node-%d", time.Now().UnixMilli())
	//fullName := fmt.Sprintf("%s.%s", NodejsTypeName, rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetPostgreSQL(context.Background(), cc, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}
				if res.Payload().Status == "TO_DELETE" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted", resource.Primary.ID)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config: fmt.Sprintf(providerBlock, org) +
				fmt.Sprintf(nodejsBlock, NodejsTypeName, rName, rName),
			Check: resource.ComposeTestCheckFunc(),
		}},
	})
}
