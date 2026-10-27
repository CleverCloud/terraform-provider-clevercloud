package dotnet_test

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
)

func TestAccDotnet_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-dotnet")
	rName2 := acctest.RandomWithPrefix("tf-test-dotnet-2")
	fullName := fmt.Sprintf("clevercloud_dotnet.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_dotnet.%s", rName2)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	dotnetBlock := helper.NewRessource(
		"clevercloud_dotnet",
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
			"environment":        map[string]any{"MY_KEY": "myval"},
			"dependencies":       []string{},
			"profile":            "dotnet-profile-1",
			"proj":               "dotnet-proj-name",
			"tfm":                "net42",
			"version":            "9.0",
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
	)
	dotnetBlock2 := helper.NewRessource(
		"clevercloud_dotnet",
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
			"repository": "https://github.com/CleverCloud/cc-dotnet-example",
			"commit":     "32d9103ab0b8719c4b172d5fa0d3b362c6a6abbb",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(dotnetBlock).String(),
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
						return tests.AssertError("expect option to be set", appRes.ForceHTTPS, "ENABLED")
					}

					if !appRes.StickySessions {
						return tests.AssertError("expect option to be set", appRes.StickySessions, true)
					}
					if appRes.Zone != "par" {
						return tests.AssertError("expect region to be 'par'", appRes.Zone, "par")
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

					v1 := env[app.CC_DOTNET_PROFILE]
					if v1 != "dotnet-profile-1" {
						return tests.AssertError("When providing 'dotnet_profile'", v1, "dotnet-profile-1")
					}

					v2 := env[app.CC_DOTNET_PROJ]
					if v2 != "dotnet-proj-name" {
						return tests.AssertError("When providing 'dotnet_proj'", v2, "dotnet-proj-name")
					}

					v3 := env[app.CC_DOTNET_TFM]
					if v3 != "net42" {
						return tests.AssertError("When providing 'dotnet_tfm'", v3, "net42")
					}

					v4 := env[app.CC_DOTNET_VERSION]
					if v4 != "9.0" {
						return tests.AssertError("When providing 'dotnet_version'", v4, "9.0")
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				dotnetBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}, {
			ResourceName: rName2,
			Config:       providerBlock.Append(dotnetBlock2).String(),
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
					vhosts := *vhostsRes.Payload()

					if len(vhosts) == 0 {
						return fmt.Errorf("there is no vhost for app: %s", id)
					}

					// Test deployed app
					err := tests.HealthCheck(ctx, application.CleverAppsVHost(vhosts, id).Fqdn, 2*time.Minute)
					if err != nil {
						return fmt.Errorf("application did not respond in the allowed time: %w", err)
					}

					return nil
				}),
			},
		}},
	})
}

func TestAccDotnet_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-dotnet-env")
	fullName := fmt.Sprintf("clevercloud_dotnet.%s", rName)
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
			"proj":               "MyApp",
			"version":            "8.0",
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_dotnet", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_DOTNET_PROJ: "Other"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_DOTNET_PROJ is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("server")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("proj"), knownvalue.StringExact("MyApp")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.APP_FOLDER:        "server",
					app.CC_DOTNET_PROJ:    "MyApp",
					app.CC_DOTNET_VERSION: "8.0",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"app_folder": "api", "proj": "Api", "version": "9.0"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("proj"), knownvalue.StringExact("Api")),
				remoteEnv(map[string]string{
					app.APP_FOLDER:        "api",
					app.CC_DOTNET_PROJ:    "Api",
					app.CC_DOTNET_VERSION: "9.0",
				}),
			},
		}},
	})
}
