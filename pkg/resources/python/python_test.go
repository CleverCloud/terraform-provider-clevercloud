package python_test

import (
	"context"
	_ "embed"
	"fmt"
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
)

func TestAccPython_basic(t *testing.T) {
	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-python")
	rName2 := acctest.RandomWithPrefix("tf-test-python-2")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_python.%s", rName2)
	vhost := "bubhbfbnriubielrbeuvieuv.com"
	cc := client.New(client.WithAutoOauthConfig())
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
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetApp(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
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
			Config:       providerBlock.Append(pythonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					if app.Name != rName {
						return assertError("invalid name", "name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return assertError("invalid min instance count", "min_instance_count", app.Instance.MinInstances, 1)
					}

					if app.Instance.MaxInstances != 2 {
						return assertError("invalid max instance count", "max_instance_count", app.Instance.MaxInstances, 2)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return assertError("invalid min flavor", "min_flavor", app.Instance.MinFlavor.Name, "XS")
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return assertError("invalid max flavor", "max_flavor", app.Instance.MaxFlavor.Name, "M")
					}

					if app.ForceHTTPS != "ENABLED" {
						return assertError("expect option to be set", "redirect_https", app.ForceHTTPS, "ENABLED")
					}

					if !app.StickySessions {
						return assertError("expect option to be set", "sticky_sessions", app.StickySessions, true)
					}

					if len(app.Vhosts) != 1 || !strings.HasSuffix(app.Vhosts[0].Fqdn, ".cleverapps.io/") {
						return assertError("invalid vhost list", "vhosts", app.Vhosts.AsString(), "1 cleverapps.io domain")
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
						return assertError("bad env var value", "env:MY_KEY", v, "myval")
					}

					v2 := env["APP_FOLDER"]
					if v2 != "./app" {
						return assertError("bad env var value", "env:APP_FOLDER", v2, "./app")
					}

					v3 := env["CC_POST_BUILD_HOOK"]
					if v3 != "echo \"build is OK!\"" {
						return assertError("bad env var value", "env:CC_POST_BUILD_HOOK", v3, "echo \"build is OK!\"")
					}

					v4 := env["CC_PIP_REQUIREMENTS_FILE"]
					if v4 != "requirements.txt" {
						return assertError("bad env var value", "env:CC_PIP_REQUIREMENTS_FILE", v4, "requirements.txt")
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
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					if len(app.Vhosts) != 1 || app.Vhosts[0].Fqdn != (vhost+"/") {
						return assertError("invalid vhost list", "vhosts", app.Vhosts.AsString(), vhost)
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

func assertError(msg, param string, got, expect any) error {
	return fmt.Errorf("%s, %s = '%v', expect: '%v'", msg, param, got, expect)
}
