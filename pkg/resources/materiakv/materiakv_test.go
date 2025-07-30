package materiakv_test

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



func TestAccMateriaKV_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-kv-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_materia_kv.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	materiakvBlock := helper.NewRessource("clevercloud_materia_kv", rName, helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetMateriaKV(ctx, cc, org, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}
				if res.Payload().Status == "TO_DELETE" || res.Payload().Status == "DELETED" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted: %+v", resource.Primary.ID, res.Payload())
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^kv_.*`)),
				resource.TestMatchResourceAttr(fullName, "host", regexp.MustCompile(`^.*clever-cloud.com$`)),
				resource.TestCheckResourceAttrSet(fullName, "port"),
				resource.TestCheckResourceAttrSet(fullName, "token"),
			),
		}},
	})
}
