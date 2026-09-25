package rust_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

func TestAccRust_basic(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-rust")
	fullName := fmt.Sprintf("clevercloud_rust.%s", rName)
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
		CheckDestroy:             tests.CheckDestroy(ctx),
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

func TestAccRust_rejectNullEnvironmentValues(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{
			{
				Config: `
resource "clevercloud_rust" "test" {
  name               = "test-rust"
  region             = "par"
  min_instance_count = 1
  max_instance_count = 1
  smallest_flavor    = "XS"
  biggest_flavor     = "XS"

  environment = {
    VAR1 = "value1"
    VAR2 = null
  }
}`,
				ExpectError: regexp.MustCompile("Null values are not allowed in environment variables"),
			},
		},
	})
}

const rustEnvSlotConfig = `
resource "clevercloud_rust" "%[1]s" {
  name               = "%[1]s"
  region             = "par"
  min_instance_count = 1
  max_instance_count = 1
  smallest_flavor    = "XS"
  biggest_flavor     = "XS"

  environment = {
    APP_FOLDER        = "server"
    CC_RUST_FEATURES  = "feature1,feature2"
    CC_PRE_BUILD_HOOK = "echo pre-build"
    MY_VAR            = "plain"
  }
}`

const rustAttrSlotConfig = `
resource "clevercloud_rust" "%[1]s" {
  name               = "%[1]s"
  region             = "par"
  min_instance_count = 1
  max_instance_count = 1
  smallest_flavor    = "XS"
  biggest_flavor     = "XS"

  app_folder = "server"
  features   = ["feature1", "feature2"]

  environment = {
    MY_VAR = "plain"
  }

  hooks {
    pre_build = "echo pre-build"
  }
}`

// TestAccRust_envSlotPreserved checks that a variable stays in the slot it was
// declared in. Each config is followed by a refresh-only step expecting an empty
// plan: that is where the bug showed, refresh moved attribute-backed variables out
// of `environment` and the next plan reverted them for ever.
func TestAccRust_envSlotPreserved(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-rust-slot")
	fullName := fmt.Sprintf("clevercloud_rust.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).String()

	envSlotChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("features"), knownvalue.Null()),
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("hooks"), knownvalue.Null()),
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.Null()),
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.MapExact(map[string]knownvalue.Check{
			"APP_FOLDER":        knownvalue.StringExact("server"),
			"CC_RUST_FEATURES":  knownvalue.StringExact("feature1,feature2"),
			"CC_PRE_BUILD_HOOK": knownvalue.StringExact("echo pre-build"),
			"MY_VAR":            knownvalue.StringExact("plain"),
		})),
	}

	attrSlotChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("features"), knownvalue.SetExact([]knownvalue.Check{
			knownvalue.StringExact("feature1"),
			knownvalue.StringExact("feature2"),
		})),
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("hooks").AtMapKey("pre_build"), knownvalue.StringExact("echo pre-build")),
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("server")),
		statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.MapExact(map[string]knownvalue.Check{
			"MY_VAR": knownvalue.StringExact("plain"),
		})),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			Config:            providerBlock + fmt.Sprintf(rustEnvSlotConfig, rName),
			ConfigStateChecks: envSlotChecks,
		}, {
			Config:             providerBlock + fmt.Sprintf(rustEnvSlotConfig, rName),
			PlanOnly:           true,
			ExpectNonEmptyPlan: false, // the fix: refresh leaves every variable in `environment`
		}, {
			Config:            providerBlock + fmt.Sprintf(rustAttrSlotConfig, rName),
			ConfigStateChecks: attrSlotChecks,
		}, {
			Config:             providerBlock + fmt.Sprintf(rustAttrSlotConfig, rName),
			PlanOnly:           true,
			ExpectNonEmptyPlan: false, // default routing still applies when `environment` is empty
		}, {
			Config:            providerBlock + fmt.Sprintf(rustEnvSlotConfig, rName),
			ConfigStateChecks: envSlotChecks,
		}, {
			Config:             providerBlock + fmt.Sprintf(rustEnvSlotConfig, rName),
			PlanOnly:           true,
			ExpectNonEmptyPlan: false, // attribute slot back to `environment` converges too
		}},
	})
}
