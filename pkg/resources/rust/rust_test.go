package rust_test

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

func TestAccRust_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-rust-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_rust.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	rustBlock := helper.NewRessource(
		"clevercloud_rust",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"build_flavor":       "M",
			"redirect_https":     true,
			"sticky_sessions":    true,
			"app_folder":         "./app",
			"environment":        map[string]any{"MY_KEY": "myval"},
			"dependencies":       []string{},
			"features":           []string{"feature1", "feature2"},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Provider,
		CheckDestroy:             testAccCheckRustDestroy(cc, org),
		Steps: []resource.TestStep{{
			Config: providerBlock.String() + rustBlock.String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "name", rName),
				resource.TestCheckResourceAttr(fullName, "region", "par"),
				resource.TestCheckResourceAttr(fullName, "min_instance_count", "1"),
				resource.TestCheckResourceAttr(fullName, "max_instance_count", "2"),
				resource.TestCheckResourceAttr(fullName, "smallest_flavor", "XS"),
				resource.TestCheckResourceAttr(fullName, "biggest_flavor", "M"),
				resource.TestCheckResourceAttr(fullName, "build_flavor", "M"),
				resource.TestCheckResourceAttr(fullName, "redirect_https", "true"),
				resource.TestCheckResourceAttr(fullName, "sticky_sessions", "true"),
				resource.TestCheckResourceAttr(fullName, "app_folder", "./app"),
				resource.TestCheckResourceAttr(fullName, "environment.MY_KEY", "myval"),
				resource.TestCheckResourceAttr(fullName, "dependencies.#", "0"),
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`app_.*`)),
				resource.TestMatchResourceAttr(fullName, "deploy_url", regexp.MustCompile(`git\+ssh://.*`)),
			),
		}, {
			Config: providerBlock.String() + helper.NewRessource(
				"clevercloud_rust",
				rName,
				helper.SetKeyValues(map[string]any{
					"name":               rName,
					"region":             "par",
					"min_instance_count": 2,
					"max_instance_count": 3,
					"smallest_flavor":    "S",
					"biggest_flavor":     "L",
					"build_flavor":       "L",
					"redirect_https":     false,
					"sticky_sessions":    false,
					"app_folder":         "./src",
					"environment":        map[string]any{"MY_KEY": "myval2", "ANOTHER_KEY": "anotherval"},
					"dependencies":       []string{},
				}),
			).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "name", rName),
				resource.TestCheckResourceAttr(fullName, "region", "par"),
				resource.TestCheckResourceAttr(fullName, "min_instance_count", "2"),
				resource.TestCheckResourceAttr(fullName, "max_instance_count", "3"),
				resource.TestCheckResourceAttr(fullName, "smallest_flavor", "S"),
				resource.TestCheckResourceAttr(fullName, "biggest_flavor", "L"),
				resource.TestCheckResourceAttr(fullName, "build_flavor", "L"),
				resource.TestCheckResourceAttr(fullName, "redirect_https", "false"),
				resource.TestCheckResourceAttr(fullName, "sticky_sessions", "false"),
				resource.TestCheckResourceAttr(fullName, "app_folder", "./src"),
				resource.TestCheckResourceAttr(fullName, "environment.MY_KEY", "myval2"),
				resource.TestCheckResourceAttr(fullName, "environment.ANOTHER_KEY", "anotherval"),
				resource.TestCheckResourceAttr(fullName, "dependencies.#", "0"),
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`app_.*`)),
				resource.TestMatchResourceAttr(fullName, "deploy_url", regexp.MustCompile(`git\+ssh://.*`)),
			),
		},
		// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCheckRustDestroy(cc *client.Client, org string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "clevercloud_rust" {
				continue
			}

			err := tmp.GetApp(context.Background(), cc, org, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Rust app %s still exists", rs.Primary.ID)
			}

			if !err.IsNotFoundError() {
				return err.Error()
			}
		}

		return nil
	}
}
