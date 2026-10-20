package python_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"strings"
	"testing"
	"time"

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
	"go.clever-cloud.dev/sdk/app"
)

func TestAccPython_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-python")
	rName2 := acctest.RandomWithPrefix("tf-test-python-2")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_python.%s", rName2)
	vhost := "bubhbfbnriubielrbeuvieuv.com"
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	pythonBlock := helper.NewRessource(
		"clevercloud_python",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"redirect_https":     true,
			"sticky_sessions":    true,
			"app_folder":         "./app",
			"python_version":     "2.7",
			"pip_requirements":   "requirements.txt",
			"environment": map[string]any{
				"MY_KEY": "myval",
			},
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(pythonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *tmp.AppResponse) error {
					if appRes.Name != rName {
						return tests.AssertError("invalid name", appRes.Name, rName)
					}

					if appRes.Instance.MinInstances != 1 {
						return tests.AssertError("invalid min instance count", appRes.Instance.MinInstances, 1)
					}

					if appRes.Instance.MaxInstances != 2 {
						return tests.AssertError("invalid max instance count", appRes.Instance.MaxInstances, 2)
					}

					if appRes.Instance.MinFlavor.Name != "XS" {
						return tests.AssertError("invalid min flavor", appRes.Instance.MinFlavor.Name, "XS")
					}

					if appRes.Instance.MaxFlavor.Name != "M" {
						return tests.AssertError("invalid max flavor", appRes.Instance.MaxFlavor.Name, "M")
					}

					if appRes.ForceHTTPS != "ENABLED" {
						return tests.AssertError("expect force_https option to be set", appRes.ForceHTTPS, "ENABLED")
					}

					if !appRes.StickySessions {
						return tests.AssertError("expect option sticky_sessions to be set", appRes.StickySessions, true)
					}

					if len(appRes.Vhosts) != 1 || !strings.HasSuffix(appRes.Vhosts[0].Fqdn, ".cleverapps.io/") {
						return tests.AssertError("invalid vhost list", appRes.Vhosts.AsString(), "1 cleverapps.io domain")
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
						return tests.AssertError("bad env var value for MY_KEY", v, "myval")
					}

					v2 := env[app.APP_FOLDER]
					if v2 != "./app" {
						return tests.AssertError("bad env var value for APP_FOLDER", v2, "./app")
					}

					v3 := env[app.CC_POST_BUILD_HOOK]
					if v3 != "echo \"build is OK!\"" {
						return tests.AssertError("bad env var value for CC_POST_BUILD_HOOK", v3, "echo \"build is OK!\"")
					}

					v4 := env["CC_PIP_REQUIREMENTS_FILE"]
					if v4 != "requirements.txt" {
						return tests.AssertError("bad env var value for CC_PIP_REQUIREMENTS_FILE", v4, "requirements.txt")
					}

					v5 := env[app.CC_PYTHON_VERSION]
					if v5 != "2.7" {
						return tests.AssertError("bad env var value for CC_PYTHON_VERSION", v5, "2.7")
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				pythonBlock.
					SetOneValue("biggest_flavor", "XS").
					SetOneValue("vhosts", []map[string]string{{"fqdn": vhost}}),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("XS")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *tmp.AppResponse) error {
					if len(appRes.Vhosts) != 1 || appRes.Vhosts[0].Fqdn != (vhost+"/") {
						return tests.AssertError("invalid vhost list", appRes.Vhosts.AsString(), vhost)
					}
					return nil
				}),
			},
		}, {
			ResourceName: rName2,
			Config: providerBlock.Append(
				helper.NewRessource(
					"clevercloud_python",
					rName2,
					helper.SetKeyValues(map[string]any{
						"name":               rName,
						"region":             "par",
						"min_instance_count": 1,
						"max_instance_count": 2,
						"smallest_flavor":    "XS",
						"biggest_flavor":     "M",
						"environment": map[string]any{
							app.CC_PYTHON_MODULE:  "app:app",
							app.CC_PYTHON_BACKEND: "gunicorn",
						},
					}),
					helper.SetBlockValues(
						"deployment",
						map[string]any{"repository": "https://github.com/CleverCloud/flask-example.git"},
					),
				),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName2, func(ctx context.Context, id string) (*tmp.VHosts, error) {
					appRes := tmp.GetAppVhosts(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, vhosts *tmp.VHosts) error {
					if len(*vhosts) == 0 {
						return fmt.Errorf("there is no vhost for app: %s", id)
					}

					err := tests.HealthCheck(ctx, vhosts.CleverAppsFQDN(id).Fqdn, 2*time.Minute)
					if err != nil {
						return fmt.Errorf("application did not respond in the allowed time: %w", err)
					}

					return nil
				}),
			},
		}},
	})
}

func TestAccPython_exposedEnv(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-python")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	pythonBlock := helper.NewRessource(
		"clevercloud_python",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"exposed_environment": map[string]any{
				"MY_EXPOSED_VAR": "initial_value",
				"ANOTHER_VAR":    "test123",
			},
		}),
	)

	pythonBlockUpdated := helper.NewRessource(
		"clevercloud_python",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"exposed_environment": map[string]any{
				"MY_EXPOSED_VAR": "updated_value",
				"NEW_VAR":        "new_value",
			},
		}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(pythonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("exposed_environment"), knownvalue.MapExact(map[string]knownvalue.Check{
					"MY_EXPOSED_VAR": knownvalue.StringExact("initial_value"),
					"ANOTHER_VAR":    knownvalue.StringExact("test123"),
				})),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					appResource := s.RootModule().Resources[fullName]
					appID := appResource.Primary.ID

					exposedEnvRes := tmp.GetExposedEnv(ctx, cc, tests.ORGANISATION, appID)
					if exposedEnvRes.HasError() {
						return fmt.Errorf("failed to get exposed env: %w", exposedEnvRes.Error())
					}

					exposedEnv := *exposedEnvRes.Payload()

					if exposedEnv["MY_EXPOSED_VAR"] != "initial_value" {
						return tests.AssertError("bad exposed env value for MY_EXPOSED_VAR", exposedEnv["MY_EXPOSED_VAR"], "initial_value")
					}

					if exposedEnv["ANOTHER_VAR"] != "test123" {
						return tests.AssertError("bad exposed env value for ANOTHER_VAR", exposedEnv["ANOTHER_VAR"], "test123")
					}

					return nil
				},
			),
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(pythonBlockUpdated).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("exposed_environment"), knownvalue.MapExact(map[string]knownvalue.Check{
					"MY_EXPOSED_VAR": knownvalue.StringExact("updated_value"),
					"NEW_VAR":        knownvalue.StringExact("new_value"),
				})),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					appResource := s.RootModule().Resources[fullName]
					appID := appResource.Primary.ID

					exposedEnvRes := tmp.GetExposedEnv(ctx, cc, tests.ORGANISATION, appID)
					if exposedEnvRes.HasError() {
						return fmt.Errorf("failed to get exposed env: %w", exposedEnvRes.Error())
					}

					exposedEnv := *exposedEnvRes.Payload()

					if exposedEnv["MY_EXPOSED_VAR"] != "updated_value" {
						return tests.AssertError("bad exposed env value for MY_EXPOSED_VAR", exposedEnv["MY_EXPOSED_VAR"], "updated_value")
					}

					if exposedEnv["NEW_VAR"] != "new_value" {
						return tests.AssertError("bad exposed env value for NEW_VAR", exposedEnv["NEW_VAR"], "new_value")
					}

					// Verify that ANOTHER_VAR was removed
					if _, exists := exposedEnv["ANOTHER_VAR"]; exists {
						return fmt.Errorf("ANOTHER_VAR should have been removed but still exists with value: %s", exposedEnv["ANOTHER_VAR"])
					}

					return nil
				},
			),
		}},
	})
}

func TestAccPython_networkgroup(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	memberFQDN := tests.NGMemberFQDN(t, "myapp")
	rName := acctest.RandomWithPrefix("tf-test-python")
	ngName := acctest.RandomWithPrefix("tf-test-python-ng")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	providerBlock2 := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	ngBlock := helper.NewRessource(
		"clevercloud_networkgroup",
		ngName,
		helper.SetKeyValues(map[string]any{
			"name":        ngName,
			"description": "for ng tests",
			"tags":        []string{"python"},
		}),
	)

	pythonBlock := helper.NewRessource(
		"clevercloud_python",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"networkgroups": []map[string]string{{
				"networkgroup_id": fmt.Sprintf("${clevercloud_networkgroup.%s.id}", ngName),
				"fqdn":            memberFQDN,
			}}}),
	)

	pythonBlock2 := helper.NewRessource(
		"clevercloud_python",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"networkgroups":      nil,
		}),
	)

	config := providerBlock.Append(ngBlock, pythonBlock)
	config2 := providerBlock2.Append(ngBlock, pythonBlock2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			// PreConfig: func() {
			// 	t.Logf("Config:\n%s", config)
			// },
			Config: config.String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("networkgroups"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"networkgroup_id": knownvalue.StringRegexp(regexp.MustCompile("^ng_.*$")),
						"fqdn":            knownvalue.StringExact(memberFQDN),
					}),
				})),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					ngResource := s.RootModule().Resources["clevercloud_networkgroup."+ngName]
					appResource := s.RootModule().Resources[fullName]
					ngID := ngResource.Primary.ID

					membersRes := tmp.ListMembers(ctx, cc, tests.ORGANISATION, ngID)
					if membersRes.HasError() {
						return fmt.Errorf("failed to list members: %w", membersRes.Error())
					}
					members := *membersRes.Payload()

					if len(members) != 1 {
						return fmt.Errorf("expect 1 member, got: %d", len(members))
					}
					member := members[0]

					if member.ID != appResource.Primary.ID {
						return fmt.Errorf("expect member to have ID %s, got: %s", appResource.Primary.ID, member.ID)
					}
					return nil
				},
			),
		}, {
			ResourceName: rName,
			// PreConfig: func() {
			// 	t.Logf("Config:\n%s", config2)
			// },
			Config: config2.String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("networkgroups"), knownvalue.Null()),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					time.Sleep(5 * time.Second) // NG API is asynchronous
					ngResource := s.RootModule().Resources["clevercloud_networkgroup."+ngName]
					ngID := ngResource.Primary.ID

					membersRes := tmp.ListMembers(ctx, cc, tests.ORGANISATION, ngID)
					if membersRes.HasError() {
						return fmt.Errorf("failed to list members: %w", membersRes.Error())
					}
					members := *membersRes.Payload()

					if len(members) != 0 {
						return fmt.Errorf("expect 0 member, got: %+v", members)
					}

					return nil
				},
			),
		}},
	})
}

func TestAccPython_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-python-env")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// config builds each step's configuration from scratch: SetOneValue
	// mutates the block, and steps must not leak values into each other.
	config := func(overrides map[string]any) string {
		values := map[string]any{
			"name":                  rName,
			"region":                "par",
			"min_instance_count":    1,
			"max_instance_count":    1,
			"smallest_flavor":       "XS",
			"biggest_flavor":        "XS",
			"backend":               "gunicorn",
			"module":                "app.server:app",
			"gunicorn_worker_class": "gevent",
			"use_gevent":            true,
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_python", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_PYTHON_MODULE: "other:app"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_PYTHON_MODULE is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("module"), knownvalue.StringExact("app.server:app")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("use_gevent"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_PYTHON_BACKEND:        "gunicorn",
					app.CC_PYTHON_MODULE:         "app.server:app",
					app.CC_GUNICORN_WORKER_CLASS: "gevent",
					app.CC_PYTHON_USE_GEVENT:     "true",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"module": "app.main:app", "gunicorn_worker_class": "gthread"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("module"), knownvalue.StringExact("app.main:app")),
				remoteEnv(map[string]string{
					app.CC_PYTHON_MODULE:         "app.main:app",
					app.CC_GUNICORN_WORKER_CLASS: "gthread",
				}),
			},
		}},
	})
}
