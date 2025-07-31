package golang_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
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

func TestAccGo_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-go-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_go.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	goBlock := helper.NewRessource(
		"clevercloud_go",
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
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
	)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetApp(ctx, cc, org, resource.Primary.ID)
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
			Config:       providerBlock.Append(goBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("M")),
				tests.NewCheckRemoteResource(
					fullName,
					func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
						appRes := tmp.GetApp(ctx, cc, org, id)
						if appRes.HasError() {
							return nil, appRes.Error()
						}
						return appRes.Payload(), nil
					},
					func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
						if app.Name != rName {
							return assertError("invalid name", app.Name, rName)
						}

						if app.Instance.MinInstances != 1 {
							return assertError("invalid min instance count", app.Instance.MinInstances, "1")
						}

						if app.Instance.MaxInstances != 2 {
							return assertError("invalid name", app.Name, rName)
						}

						if app.Instance.MinFlavor.Name != "XS" {
							return assertError("invalid name", app.Name, rName)
						}

						if app.Instance.MaxFlavor.Name != "M" {
							return assertError("invalid name", app.Name, rName)
						}

						if app.ForceHTTPS != "ENABLED" {
							return assertError("expect option to be set", "redirect_https", app.ForceHTTPS)
						}

						if !app.StickySessions {
							return assertError("expect option to be set", "sticky_sessions", app.StickySessions)
						}
						if app.Zone != "par" {
							return assertError("expect region to be 'par'", "region", app.Zone)
						}

						appEnvRes := tmp.GetAppEnv(ctx, cc, org, id)
						if appEnvRes.HasError() {
							return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
						}

						env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
							acc[e.Name] = e.Value
							return acc
						})

						v := env["MY_KEY"]
						if v != "myval" {
							return assertError("bad env var value MY_KEY", "myval3", v)
						}

						v2 := env["APP_FOLDER"]
						if v2 != "./app" {
							return assertError("bad env var value APP_FOLER", "./app", v2)
						}

						v3 := env["CC_POST_BUILD_HOOK"]
						if v3 != "echo \"build is OK!\"" {
							return assertError("bad env var value CC_POST_BUILD_HOOK", "echo \"build is OK!\"", v3)
						}
						return nil
					}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				goBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}},
	})
}

func assertError(msg string, a, b any) error {
	return fmt.Errorf("%s, got: '%v', expect: '%v'", msg, a, b)
}
