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

//go:embed resource_postgresql_test_block.tf
var postgresqlBlock string

func TestAccPostgreSQL_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-pg-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
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
			Config:       fmt.Sprintf(providerBlock, org) + fmt.Sprintf(postgresqlBlock, rName, rName),
			Check: resource.ComposeTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^addon_.*`)),
				resource.TestMatchResourceAttr(fullName, "host", regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`)),
				resource.TestCheckResourceAttrSet(fullName, "port"),
				resource.TestMatchResourceAttr(fullName, "database", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestMatchResourceAttr(fullName, "user", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestMatchResourceAttr(fullName, "password", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
			),
		}},
	})
}
