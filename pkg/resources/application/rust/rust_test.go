package rust_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/app"
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

func TestAccRust_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-rust-env")
	fullName := fmt.Sprintf("clevercloud_rust.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// config builds each step's configuration from scratch: SetOneValue
	// mutates the block, and steps must not leak values into each other.
	config := func(overrides map[string]any) string {
		values := map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"features":           []string{"tls", "metrics"},
			"rustup_channel":     "nightly",
			"bin":                "server",
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_rust", rName, helper.SetKeyValues(values))).String()
	}

	remoteEnv := func(want map[string]string) statecheck.StateCheck {
		return tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
			appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
			if appRes.HasError() {
				return nil, appRes.Error()
			}
			return appRes.Payload(), nil
		}, func(ctx context.Context, id string, state *tfjson.State, appRes *tmp.AppResponse) error {
			appEnvRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
			if appEnvRes.HasError() {
				return fmt.Errorf("failed to get application env: %w", appEnvRes.Error())
			}
			env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
				acc[e.Name] = e.Value
				return acc
			})
			for name, value := range want {
				if env[name] != value {
					return tests.AssertError("bad env var value "+name, value, env[name])
				}
			}
			return nil
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// First, so the post-test destroy runs with a valid configuration.
			ResourceName: rName,
			Config:       config(map[string]any{"environment": map[string]any{app.CC_RUSTUP_CHANNEL: "beta"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_RUSTUP_CHANNEL is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("features"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("tls"),
					knownvalue.StringExact("metrics"),
				})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("rustup_channel"), knownvalue.StringExact("nightly")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("bin"), knownvalue.StringExact("server")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_RUSTUP_CHANNEL: "nightly",
					app.CC_RUST_BIN:       "server",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"features": []string{"tls"}, "rustup_channel": "1.85"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("features"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("tls"),
				})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("rustup_channel"), knownvalue.StringExact("1.85")),
				remoteEnv(map[string]string{
					app.CC_RUST_FEATURES:  "tls",
					app.CC_RUSTUP_CHANNEL: "1.85",
				}),
			},
		}},
	})
}
