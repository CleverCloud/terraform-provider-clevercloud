package postgresql_test

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

func TestAccPostgreSQL_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-pg-%d", time.Now().UnixMilli())
	rName2 := fmt.Sprintf("tf-test2-pg-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_postgresql.%s", rName2)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "dev",
		}))

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
			Config:       providerBlock.Append(postgresqlBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^addon_.*`)),
				resource.TestMatchResourceAttr(fullName, "host", regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`)),
				resource.TestCheckResourceAttrSet(fullName, "port"),
				resource.TestMatchResourceAttr(fullName, "database", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestMatchResourceAttr(fullName, "user", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestMatchResourceAttr(fullName, "password", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestCheckResourceAttr(fullName, "plan", "dev"),
			),
		}, {
			ResourceName: rName2,
			Config: providerBlock.Append(helper.NewRessource(
				"clevercloud_postgresql",
				rName2,
				helper.SetKeyValues(map[string]any{
					"name":    rName2,
					"region":  "par",
					"plan":    "xs_sml",
					"version": "17",
				}))).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName2, "id", regexp.MustCompile(`^addon_.*`)),
				resource.TestMatchResourceAttr(fullName2, "host", regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`)),
				resource.TestCheckResourceAttrSet(fullName2, "port"),
				resource.TestMatchResourceAttr(fullName2, "database", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestMatchResourceAttr(fullName2, "user", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestMatchResourceAttr(fullName2, "password", regexp.MustCompile(`^[a-zA-Z0-9]+$`)),
				resource.TestCheckResourceAttr(fullName2, "plan", "xs_sml"),
				resource.TestCheckResourceAttr(fullName2, "version", "17"),
			),
		}, /*{
			ResourceName: rName3,
			Config: providerBlock.Append(helper.NewRessource(
				"clevercloud_postgresql",
				rName3,
				helper.SetKeyValues(map[string]any{
					"name":    rName2,
					"region":  "par",
					"plan":    "dev",
					"version": "10",
				}))).String(),
			ExpectError: regexp.MustCompile(`Error running pre-apply plan`),
		},*/
		},
	})
}

func TestAccPostgreSQL_RefreshDeleted(t *testing.T) {
	rName := fmt.Sprintf("tf-test-pg-%d", time.Now().UnixMilli())
	//fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "dev",
		}))

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
		Steps: []resource.TestStep{
			// create a database instance on first step
			{
				ResourceName: rName,
				Config:       providerBlock.Append(postgresqlBlock).String(),
			},
			{
				ResourceName: rName,
				PreConfig: func() {
					// delete the database using an api call
					tmp.DeleteAddon(context.Background(), cc, org, rName)
				},
				// refreshing state
				RefreshState: true,
				// plan should contain database re-creation
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
