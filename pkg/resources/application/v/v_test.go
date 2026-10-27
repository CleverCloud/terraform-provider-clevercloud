package v_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"testing"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/application"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/app"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/services/applications"
)

func TestAccV_basic(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-v")
	rName2 := acctest.RandomWithPrefix("tf-test-v-2")
	fullName := fmt.Sprintf("clevercloud_v.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_v.%s", rName2)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	vBlock := helper.NewRessource(
		"clevercloud_v",
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
			"binary":             "a.out",
			"development_build":  true,
		}),
	)
	vBlock2 := helper.NewRessource(
		"clevercloud_v",
		rName2,
		helper.SetKeyValues(map[string]any{
			"name":               rName2,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
		}),
		helper.SetBlockValues("deployment", map[string]any{
			"repository": "https://github.com/CleverCloud/v-example",
			"commit":     "fc8a99ad4f10a9aaed1103dcf94f3a4dcecc4203",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(vBlock).String(),
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

					v1 := env[app.CC_V_BINARY]
					if v1 != "a.out" {
						return tests.AssertError("When providing 'v_binary': 'a.out'", v1, "a.out")
					}

					v2 := env[app.ENVIRONMENT]
					if v2 != "development" {
						return tests.AssertError("When providing 'v_environment': 'development'", v2, "development")
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				vBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}, {
			ResourceName: rName2,
			Config:       providerBlock.Append(vBlock2).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName2, func(ctx context.Context, id string) (*models.ApplicationView, error) {
					appRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Getorganisationapplication(ctx)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *models.ApplicationView) error {
					vhostsRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Vhosts().Listorganisationapplicationvirtualhosts(ctx)
					if vhostsRes.HasError() {
						return fmt.Errorf("failed to get application vhosts: %w", vhostsRes.Error())
					}
					vhosts := vhostsRes.Payload()

					if len(*vhosts) == 0 {
						return fmt.Errorf("there is no vhost for app: %s", id)
					}

					// Test deployed app
					err := tests.HealthCheck(ctx, application.CleverAppsVHost(*vhosts, id).Fqdn, 2*time.Minute)
					if err != nil {
						return fmt.Errorf("application did not respond in the allowed time: %w", err)
					}

					return nil
				}),
			},
		}},
	})
}

// TestAccV_EnvironmentDriftDetection reproduces issue #343
// This test verifies that environment variable changes made through the console
// (simulated via API) are properly detected by Terraform during refresh/plan operations.
func TestAccV_EnvironmentDriftDetection(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-v")
	fullName := fmt.Sprintf("clevercloud_v.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	var appID string

	// Initial configuration with environment variable
	vBlock := helper.NewRessource(
		"clevercloud_v",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"environment": map[string]any{
				"MY_VAR":      "initial_value",
				"ANOTHER_VAR": "stable_value",
			},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(vBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment").AtMapKey("MY_VAR"), knownvalue.StringExact("initial_value")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment").AtMapKey("ANOTHER_VAR"), knownvalue.StringExact("stable_value")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*models.ApplicationView, error) {
					appRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Getorganisationapplication(ctx)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *models.ApplicationView) error {
					// Verify environment variables are set correctly in the API
					appEnvRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Env().Listorganisationapplicationenvironmentvariables(ctx)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application env: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e models.AddonEnvironmentView) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if env["MY_VAR"] != "initial_value" {
						return tests.AssertError("bad env var value MY_VAR", env["MY_VAR"], "initial_value")
					}
					if env["ANOTHER_VAR"] != "stable_value" {
						return tests.AssertError("bad env var value ANOTHER_VAR", env["ANOTHER_VAR"], "stable_value")
					}

					return nil
				}),
			},
			Check: resource.TestCheckResourceAttrWith(fullName, "id", func(value string) error {
				appID = value
				return nil
			}),
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				helper.NewRessource(
					"clevercloud_v",
					rName,
					helper.SetKeyValues(map[string]any{
						"name":               rName,
						"region":             "par",
						"min_instance_count": 1,
						"max_instance_count": 1,
						"smallest_flavor":    "XS",
						"biggest_flavor":     "XS",
						"environment": map[string]any{
							"MY_VAR":      "console_modified",
							"ANOTHER_VAR": "stable_value",
						},
					}),
				),
			).String(),
			PreConfig: func() {
				// Simulate a console modification: change MY_VAR to "console_modified"
				// This simulates what happens when a user modifies env vars through the Clever Cloud console
				envUpdate := map[string]string{
					"MY_VAR":      "console_modified",
					"ANOTHER_VAR": "stable_value",
				}
				envRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(appID).Env().Updateorganisationapplicationenvironmentvariables(ctx, envUpdate, applications.WithCheck(true))
				if envRes.HasError() {
					t.Fatalf("failed to update env via API: %v", envRes.Error())
				}
			},
			ConfigStateChecks: []statecheck.StateCheck{
				// EXPECTED BEHAVIOR (with fix): State reflects API reality after read
				// When Terraform reads the resource, it should see "console_modified" from the API
				// and store it in state, making the plan empty (no changes needed)
				// Without the fix for issue #343, the state would still show "initial_value"
				// and Terraform would plan to change it to "console_modified"
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment").AtMapKey("MY_VAR"), knownvalue.StringExact("console_modified")),
			},
		},
		},
	})
}

func TestAccV_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-v-env")
	fullName := fmt.Sprintf("clevercloud_v.%s", rName)
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
			"app_folder":         "./app",
			"binary":             "bin/server",
			"development_build":  true,
			"modules_path":       "/home/bas/modules",
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_v", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.VMODULES: "/tmp/modules"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`VMODULES is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("./app")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("development_build"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("modules_path"), knownvalue.StringExact("/home/bas/modules")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.APP_FOLDER:  "./app",
					app.CC_V_BINARY: "bin/server",
					app.ENVIRONMENT: "development",
					app.VMODULES:    "/home/bas/modules",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"development_build": false, "modules_path": "/home/bas/vendor"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("development_build"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("modules_path"), knownvalue.StringExact("/home/bas/vendor")),
				remoteEnv(map[string]string{
					app.ENVIRONMENT: "",
					app.VMODULES:    "/home/bas/vendor",
				}),
			},
		}},
	})
}
