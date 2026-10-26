package linux_test

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
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/app"
	"go.clever-cloud.dev/sdk/models"
)

func TestAccLinux_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-linux")
	fullName := fmt.Sprintf("clevercloud_linux.%s", rName)
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
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(linuxBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("XL")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*models.ApplicationView, error) {
					appRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Getorganisationapplication(ctx)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *models.ApplicationView) error {
					if appRes.Name != rName {
						return tests.AssertError("invalid name", appRes.Name, rName)
					}

					if appRes.Instance.MinInstances != 1 {
						return tests.AssertError("invalid min instance count", appRes.Instance.MinInstances, "1")
					}

					if appRes.Instance.MaxInstances != 2 {
						return tests.AssertError("invalid name", appRes.Instance.MaxInstances, 2)
					}

					if pkg.FlavorName(appRes.Instance.MinFlavor) != "XS" {
						return tests.AssertError("invalid name", pkg.FlavorName(appRes.Instance.MinFlavor), "XS")
					}

					if pkg.FlavorName(appRes.Instance.MaxFlavor) != "M" {
						return tests.AssertError("invalid max instance name", pkg.FlavorName(appRes.Instance.MaxFlavor), "M")
					}

					if pkg.FlavorName(appRes.BuildFlavor) != "XL" {
						return tests.AssertError("invalid build flavor", pkg.FlavorName(appRes.BuildFlavor), "XL")
					}

					if appRes.ForceHTTPS != "ENABLED" {
						return tests.AssertError("expect option to be set", "redirect_https", appRes.ForceHTTPS)
					}

					if !appRes.StickySessions {
						return tests.AssertError("expect option to be set", "sticky_sessions", appRes.StickySessions)
					}
					if appRes.Zone != "par" {
						return tests.AssertError("expect region to be 'par'", "region", appRes.Zone)
					}
					appEnvRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Env().Listorganisationapplicationenvironmentvariables(ctx)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e models.AddonEnvironmentView) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					v := env["MY_KEY"]
					if v != "myval" {
						return tests.AssertError("bad env var value MY_KEY", v, "myval")
					}

					v1 := env[app.CC_RUN_COMMAND]
					if v1 != "./start.sh" {
						return tests.AssertError("When providing 'run_command': './start.sh'", v1, "./start.sh")
					}

					v2 := env[app.CC_BUILD_COMMAND]
					if v2 != "make build" {
						return tests.AssertError("When providing 'build_command': 'make build'", v2, "make build")
					}

					v3 := env["CC_MAKEFILE"]
					if v3 != "Makefile.custom" {
						return tests.AssertError("When providing 'makefile': 'Makefile.custom'", v3, "Makefile.custom")
					}

					v4 := env[app.CC_MISE_FILE_PATH]
					if v4 != "./tools/mise.toml" {
						return tests.AssertError("When providing 'mise_file_path': './tools/mise.toml'", v4, "./tools/mise.toml")
					}

					v5 := env[app.CC_DISABLE_MISE]
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

func TestAccLinux_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-linux-env")
	fullName := fmt.Sprintf("clevercloud_linux.%s", rName)
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
			"app_folder":         "server",
			"build_command":      "make build",
			"makefile":           "Makefile.custom",
			"disable_mise":       true,
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_linux", rName, helper.SetKeyValues(values))).String()
	}

	remoteEnv := func(want map[string]string) statecheck.StateCheck {
		return tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*models.ApplicationView, error) {
			appRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Getorganisationapplication(ctx)
			if appRes.HasError() {
				return nil, appRes.Error()
			}
			return appRes.Payload(), nil
		}, func(ctx context.Context, id string, state *tfjson.State, appRes *models.ApplicationView) error {
			appEnvRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Env().Listorganisationapplicationenvironmentvariables(ctx)
			if appEnvRes.HasError() {
				return fmt.Errorf("failed to get application env: %w", appEnvRes.Error())
			}
			env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e models.AddonEnvironmentView) map[string]string {
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_BUILD_COMMAND: "make other"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_BUILD_COMMAND is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("server")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_command"), knownvalue.StringExact("make build")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("disable_mise"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.APP_FOLDER:       "server",
					app.CC_BUILD_COMMAND: "make build",
					"CC_MAKEFILE":        "Makefile.custom",
					app.CC_DISABLE_MISE:  "true",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"app_folder": "api", "build_command": "make api", "disable_mise": false}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_command"), knownvalue.StringExact("make api")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("disable_mise"), knownvalue.Bool(false)),
				remoteEnv(map[string]string{
					app.APP_FOLDER:       "api",
					app.CC_BUILD_COMMAND: "make api",
					app.CC_DISABLE_MISE:  "",
				}),
			},
		}},
	})
}
