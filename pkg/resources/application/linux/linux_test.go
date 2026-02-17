package linux_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccLinux_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-linux")
	fullName := fmt.Sprintf("clevercloud_linux.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	linuxBlock := helper.NewRessource(
		"clevercloud_linux",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"build_flavor":       "XL",
			"redirect_https":     true,
			"sticky_sessions":    true,
			"app_folder":         "./app",
			"environment":        map[string]any{"MY_KEY": "myval"},
			"dependencies":       []string{},
			"run_command":        "./start.sh",
			"build_command":      "make build",
			"makefile":           "Makefile.custom",
			"mise_file_path":     "./tools/mise.toml",
			"disable_mise":       true,
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetApp(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpected error: %s", res.Error().Error())
				}
				if res.Payload().State == "TO_DELETE" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted state: '%s'", resource.Primary.ID, res.Payload().State)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(linuxBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("XL")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.AppResponse) error {
					if app.Name != rName {
						return tests.AssertError("invalid name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return tests.AssertError("invalid min instance count", app.Instance.MinInstances, "1")
					}

					if app.Instance.MaxInstances != 2 {
						return tests.AssertError("invalid name", app.Instance.MaxInstances, 2)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return tests.AssertError("invalid name", app.Instance.MinFlavor.Name, "XS")
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return tests.AssertError("invalid max instance name", app.Instance.MaxFlavor.Name, "M")
					}

					if app.BuildFlavor.Name != "XL" {
						return tests.AssertError("invalid build flavor", app.BuildFlavor.Name, "XL")
					}

					if app.ForceHTTPS != "ENABLED" {
						return tests.AssertError("expect option to be set", "redirect_https", app.ForceHTTPS)
					}

					if !app.StickySessions {
						return tests.AssertError("expect option to be set", "sticky_sessions", app.StickySessions)
					}
					if app.Zone != "par" {
						return tests.AssertError("expect region to be 'par'", "region", app.Zone)
					}
					appEnvRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					v := env["MY_KEY"]
					if v != "myval" {
						return tests.AssertError("bad env var value MY_KEY", v, "myval")
					}

					v1 := env["CC_RUN_COMMAND"]
					if v1 != "./start.sh" {
						return tests.AssertError("When providing 'run_command': './start.sh'", v1, "./start.sh")
					}

					v2 := env["CC_BUILD_COMMAND"]
					if v2 != "make build" {
						return tests.AssertError("When providing 'build_command': 'make build'", v2, "make build")
					}

					v3 := env["CC_MAKEFILE"]
					if v3 != "Makefile.custom" {
						return tests.AssertError("When providing 'makefile': 'Makefile.custom'", v3, "Makefile.custom")
					}

					v4 := env["CC_MISE_FILE_PATH"]
					if v4 != "./tools/mise.toml" {
						return tests.AssertError("When providing 'mise_file_path': './tools/mise.toml'", v4, "./tools/mise.toml")
					}

					v5 := env["CC_DISABLE_MISE"]
					if v5 != "true" {
						return tests.AssertError("When providing 'disable_mise': true", v5, "true")
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				linuxBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}},
	})
}
