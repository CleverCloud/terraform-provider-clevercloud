package frankenphp_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"strings"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
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

func TestAccFrankenPHP_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-frankenphp")
	fullName := fmt.Sprintf("clevercloud_frankenphp.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	domain := fmt.Sprintf("%s.com", rName)
	domainEdit1 := rName + "-1.com"
	domainEdit2 := rName + "-2.com"
	frankenphpBlock := helper.NewRessource(
		"clevercloud_frankenphp",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"dev_dependencies":   false,
			"vhosts":             []map[string]string{{"fqdn": domain}},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(frankenphpBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("dev_dependencies"), knownvalue.Bool(false)),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.AppResponse) error {
					if len(app.Vhosts) != 1 || app.Vhosts[0].Fqdn != (domain+"/") {
						return tests.AssertError("invalid vhost list", app.Vhosts.AsString(), "1 cleverapps.io domain")
					}
					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				frankenphpBlock.
					SetOneValue("min_instance_count", 2).
					SetOneValue("max_instance_count", 6).
					SetOneValue("dev_dependencies", true).
					SetOneValue("vhosts", []map[string]string{{"fqdn": domainEdit1}, {"fqdn": domainEdit2, "path_begin": "/test"}}),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("dev_dependencies"), knownvalue.Bool(true)),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.AppResponse) error {
					expectedVhosts := []string{domainEdit1 + "/", domainEdit2 + "/test"}
					if len(app.Vhosts) != len(expectedVhosts) {
						return tests.AssertError("invalid vhost count", len(app.Vhosts), len(expectedVhosts))
					}

					for _, expectedVhost := range expectedVhosts {
						found := false
						for _, vhost := range app.Vhosts {
							if vhost.Fqdn == expectedVhost {
								found = true
								break
							}
						}
						if !found {
							return tests.AssertError("missing expected vhost", app.Vhosts.AsString(), expectedVhost)
						}
					}

					for _, vhost := range app.Vhosts {
						if strings.HasSuffix(vhost.Fqdn, ".cleverapps.io/") {
							return tests.AssertError("cleverapps.io should not be present with custom vhosts", app.Vhosts.AsString(), "no cleverapps.io domain")
						}
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				frankenphpBlock.
					SetOneValue("min_instance_count", 2).
					SetOneValue("max_instance_count", 6).
					SetOneValue("dev_dependencies", true).
					UnsetOneValue("vhosts"),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("dev_dependencies"), knownvalue.Bool(true)),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.AppResponse) error {
					if len(app.Vhosts) != 2 {
						return tests.AssertError("invalid vhost list", app.Vhosts.AsString(), "2 custom domain")
					}
					return nil
				}),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccFrankenPHP_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-frankenphp-env")
	fullName := fmt.Sprintf("clevercloud_frankenphp.%s", rName)
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
			"app_folder":         "app",
			"webroot":            "/public",
			"worker_script":      "/public/worker.php",
			"port":               8081,
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_frankenphp", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_FRANKENPHP_WORKER: "/other.php"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_FRANKENPHP_WORKER is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("app")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.Int64Exact(8081)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.APP_FOLDER:           "app",
					app.CC_WEBROOT:           "/public",
					app.CC_FRANKENPHP_WORKER: "/public/worker.php",
					app.CC_FRANKENPHP_PORT:   "8081",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"app_folder": "src", "port": 9000, "composer_flags": "-n --no-progress"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("app_folder"), knownvalue.StringExact("src")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.Int64Exact(9000)),
				remoteEnv(map[string]string{
					app.APP_FOLDER:            "src",
					app.CC_FRANKENPHP_PORT:    "9000",
					app.CC_PHP_COMPOSER_FLAGS: "-n --no-progress",
				}),
			},
		}},
	})
}
