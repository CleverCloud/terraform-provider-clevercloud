package rust_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccRust_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-rust")
	fullName := fmt.Sprintf("clevercloud_rust.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
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
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             testAccCheckRustDestroy(cc, tests.ORGANISATION),
		Steps: []resource.TestStep{{
			Config: providerBlock.String() + rustBlock.String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(1)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("smallest_flavor"), knownvalue.StringExact("XS")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("M")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("M")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("redirect_https"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("sticky_sessions"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("./app")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment").AtMapKey("MY_KEY"), knownvalue.StringExact("myval")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("dependencies"), knownvalue.ListSizeExact(0)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`app_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`git\+ssh://.*`))),
			},
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
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(3)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("smallest_flavor"), knownvalue.StringExact("S")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("L")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("L")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("redirect_https"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("sticky_sessions"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("./src")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment").AtMapKey("MY_KEY"), knownvalue.StringExact("myval2")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment").AtMapKey("ANOTHER_KEY"), knownvalue.StringExact("anotherval")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("dependencies"), knownvalue.ListSizeExact(0)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`app_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`git\+ssh://.*`))),
			},
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
