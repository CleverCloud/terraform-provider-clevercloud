package haskell_test

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"regexp"
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

func TestAccNodejs_basic(t *testing.T) {
	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-haskell")
	rName2 := acctest.RandomWithPrefix("tf-test-haskell-2")
	fullName := fmt.Sprintf("clevercloud_haskell.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_haskell.%s", rName2)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	haskellBlock := helper.NewRessource(
		"clevercloud_haskell",
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
		}),
	)
	haskellBlock2 := helper.NewRessource(
		"clevercloud_haskell",
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
			"repository": "https://github.com/CleverCloud/haskell-scotty-example",
			"commit":     "fe61c758a1dceeb9fe1cfce6d6c9dd91f6bb677f",
		}))

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
			Config:       providerBlock.Append(haskellBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("XL")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					if app.Name != rName {
						return assertError("invalid name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return assertError("invalid min instance count", app.Instance.MinInstances, "1")
					}

					if app.Instance.MaxInstances != 2 {
						return assertError("invalid name", app.Instance.MaxInstances, 2)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return assertError("invalid name", app.Instance.MinFlavor.Name, "XS")
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return assertError("invalid max instance name", app.Instance.MaxFlavor.Name, "M")
					}

					if app.BuildFlavor.Name != "XL" {
						return assertError("invalid build flavor", app.BuildFlavor.Name, "XL")
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
						return assertError("bad env var value MY_KEY", "myval3", v)
					}
					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				haskellBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}, {
			ResourceName: rName2,
			Config:       providerBlock.Append(haskellBlock2).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName2, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					vhostsRes := tmp.GetAppVhosts(ctx, cc, tests.ORGANISATION, id)
					if vhostsRes.HasError() {
						return fmt.Errorf("failed to get application vhosts: %w", vhostsRes.Error())
					}
					vhosts := vhostsRes.Payload()

					if len(*vhosts) == 0 {
						return fmt.Errorf("there is no vhost for app: %s", id)
					}

					// Test deployed app
					t := time.NewTimer(15 * time.Minute)
					select {
					case <-healthCheck(vhosts.CleverAppsFQDN(id).Fqdn):
						return nil
					case <-t.C:
						return fmt.Errorf("application did not respond in the allowed time")
					}
				}),
			},
		}},
	})
}

func assertError(msg string, a, b any) error {
	return fmt.Errorf("%s, got: '%v', expect: '%v'", msg, a, b)
}

func healthCheck(vhost string) chan struct{} {
	c := make(chan struct{})

	fmt.Printf("Test on %s\n", vhost)

	go func() {
		for {
			res, err := http.Get(fmt.Sprintf("https://%s", vhost))
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}

			fmt.Printf("RESPONSE %d\n", res.StatusCode)
			if res.StatusCode != 200 {
				time.Sleep(1 * time.Second)
				continue
			}
			c <- struct{}{}
		}
	}()

	return c
}
