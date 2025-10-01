package configprovider_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.dev/client"
)

func TestAccConfigProvider_basic(t *testing.T) {
	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-cp")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_configprovider.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	configProviderBlock := helper.NewRessource(
		"clevercloud_configprovider",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par", "plan": "std", "environment": map[string]any{"foo": "this is foo"}}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				addonId, err := tmp.RealIDToAddonID(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						continue
					}
					return fmt.Errorf("failed to get addon ID: %s", err.Error())
				}

				res := tmp.GetConfigProvider(ctx, cc, addonId)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}
				if res.Payload().Status == "TO_DELETE" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted", resource.Primary.ID)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(configProviderBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^config_.*`))),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.ConfigProvider, error) {
					addonId, err := tmp.RealIDToAddonID(ctx, cc, tests.ORGANISATION, id)
					if err != nil {
						return nil, fmt.Errorf("failed to get addon ID: %s", err.Error())
					}
					res := tmp.GetConfigProvider(ctx, cc, addonId)
					if res.IsNotFoundError() {
						return nil, fmt.Errorf("Unable to find configProvider by real id " + addonId)
					}
					if res.HasError() {
						return nil, fmt.Errorf("unexpectd error: %s", res.Error().Error())
					}
					return res.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.ConfigProvider) error {
					// Verify instance counts were updated

					// Verify environment variables were updated
					cpEnvRes := tmp.GetConfigProviderEnv(ctx, cc, tests.ORGANISATION, id)
					if cpEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", cpEnvRes.Error())
					}

					env := pkg.Reduce(*cpEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if v := env["MY_KEY"]; v != "updated_val" {
						return assertError("bad updated env var value MY_KEY", "updated_val", v)
					}


					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(configProviderBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^config_.*`))),
			},
		}},
	})
}
// assertError helper function for cleaner error assertions
func assertError(message string, got, want any) error {
	return fmt.Errorf("%s: got %v, want %v", message, got, want)
}
