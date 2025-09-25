package frankenphp_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccFrankenPHP_basic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-frankenphp")
	fullName := fmt.Sprintf("clevercloud_frankenphp.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	domain := fmt.Sprintf("%s.com", rName)
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
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					if len(app.Vhosts) != 1 || app.Vhosts[0].Fqdn != (domain+"/") {
						return assertError("invalid vhost list", "vhosts", app.Vhosts.AsString(), "1 cleverapps.io domain")
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
					SetOneValue("vhosts", []map[string]string{{"fqdn": "test-frankenphp-1.com"}, {"fqdn": "test-frankenphp-2.com", "path_begin": "/test"}}),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("dev_dependencies"), knownvalue.Bool(true)),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					expectedVhosts := []string{"test-frankenphp-1.com/", "test-frankenphp-2.com/test"}
					if len(app.Vhosts) != len(expectedVhosts) {
						return assertError("invalid vhost count", "vhosts", len(app.Vhosts), len(expectedVhosts))
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
							return assertError("missing expected vhost", "vhosts", app.Vhosts.AsString(), expectedVhost)
						}
					}

					for _, vhost := range app.Vhosts {
						if strings.HasSuffix(vhost.Fqdn, ".cleverapps.io/") {
							return assertError("cleverapps.io should not be present with custom vhosts", "vhosts", app.Vhosts.AsString(), "no cleverapps.io domain")
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
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					if len(app.Vhosts) != 2 {
						return assertError("invalid vhost list", "vhosts", app.Vhosts.AsString(), "2 custom domain")
					}
					return nil
				}),
			},
		}},
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
	})
}

func assertError(msg, param string, got, expect any) error {
	return fmt.Errorf("%s, %s = '%v', expect: '%v'", msg, param, got, expect)
}
