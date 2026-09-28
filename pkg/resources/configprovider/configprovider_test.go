package configprovider_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

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

func TestAccConfigProvider_basic(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-cp")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_configprovider.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	configProviderBlock := helper.NewRessource(
		"clevercloud_configprovider",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "environment": map[string]any{"foo": "this is foo"}}),
	)

	retrieveConfigProvider := func(ctx context.Context, id string) (*tmp.ConfigProvider, error) {
		res := tmp.GetConfigProvider(ctx, cc, id)
		if res.IsNotFoundError() {
			return nil, fmt.Errorf("Unable to find configProvider by real id: %s", id)
		}
		if res.HasError() {
			return nil, fmt.Errorf("Unexpectd error: %s", res.Error().Error())
		}
		return res.Payload(), nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(configProviderBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^config_.*`))),
				tests.NewCheckRemoteResource(fullName, retrieveConfigProvider, func(ctx context.Context, id string, state *tfjson.State, app *tmp.ConfigProvider) error {
					// Verify environment variables were updated
					cpEnvRes := tmp.GetConfigProviderEnv(ctx, cc, tests.ORGANISATION, id)
					if cpEnvRes.HasError() {
						return fmt.Errorf("Failed to get application: %w", cpEnvRes.Error())
					}

					env := pkg.Reduce(*cpEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.EnvVar) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if len(env) != 1 {
						return tests.AssertError("Env should only have 1 value", env, "env:foo")
					}

					if v := env["foo"]; v != "this is foo" {
						return tests.AssertError("Bad updated env var value MY_KEY", v, "this is foo")
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
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(configProviderBlock.SetOneValue("environment", map[string]any{"foo": "this is foo", "bar": "this is bar"})).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName, retrieveConfigProvider, func(ctx context.Context, id string, state *tfjson.State, app *tmp.ConfigProvider) error {
					cpEnvRes := tmp.GetConfigProviderEnv(ctx, cc, tests.ORGANISATION, id)
					if cpEnvRes.HasError() {
						return fmt.Errorf("Failed to get application: %w", cpEnvRes.Error())
					}

					env := pkg.Reduce(*cpEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.EnvVar) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if len(env) != 2 {
						return tests.AssertError("Env should only have 1 value", env, "env:foo")
					}

					if v := env["foo"]; v != "this is foo" {
						return tests.AssertError("Bad updated env foo", v, "this is foo")
					}

					if v2 := env["bar"]; v2 != "this is bar" {
						return tests.AssertError("Bad updated env bar", v2, "this is bar")
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(configProviderBlock.SetOneValue("environment", map[string]any{"bar": "this is bar"})).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName, retrieveConfigProvider, func(ctx context.Context, id string, state *tfjson.State, app *tmp.ConfigProvider) error {
					// Verify environment variables were updated
					cpEnvRes := tmp.GetConfigProviderEnv(ctx, cc, tests.ORGANISATION, id)
					if cpEnvRes.HasError() {
						return fmt.Errorf("Failed to get application: %w", cpEnvRes.Error())
					}

					env := pkg.Reduce(*cpEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.EnvVar) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if len(env) != 1 {
						return tests.AssertError("Env should only have 1 value", env, "env:foo")
					}

					if v2 := env["bar"]; v2 != "this is bar" {
						return tests.AssertError("Bad updated env bar", v2, "this is bar")
					}

					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(configProviderBlock.SetOneValue("environment", map[string]any{})).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName, retrieveConfigProvider, func(ctx context.Context, id string, state *tfjson.State, app *tmp.ConfigProvider) error {
					// Verify environment variables were updated
					cpEnvRes := tmp.GetConfigProviderEnv(ctx, cc, tests.ORGANISATION, id)
					if cpEnvRes.HasError() {
						return fmt.Errorf("Failed to get application: %w", cpEnvRes.Error())
					}

					env := pkg.Reduce(*cpEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.EnvVar) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if len(env) != 0 {
						return tests.AssertError("Env should have 0 entries", env, "[]")
					}

					return nil
				}),
			},
		}},
	})
}

func TestAccConfigProvider_Import(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-cp-import")
	fullName := fmt.Sprintf("clevercloud_configprovider.%s", rName)

	var addonID string
	var realID string

	// Cleanup addon created via API in case the import step fails
	// before Terraform takes ownership of the resource.
	t.Cleanup(func() {
		if addonID != "" {
			tmp.DeleteAddon(context.Background(), cc, tests.ORGANISATION, addonID)
		}
	})

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	configProviderBlock := helper.NewRessource(
		"clevercloud_configprovider",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "environment": map[string]any{"foo": "this is foo"}}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create the add-on through the API to simulate a pre-existing
				// resource, then import it into Terraform state.
				PreConfig: func() {
					addonsProvidersRes := tmp.GetAddonsProviders(ctx, cc)
					if addonsProvidersRes.HasError() {
						t.Fatalf("failed to get addon providers: %s", addonsProvidersRes.Error())
					}
					prov := pkg.LookupAddonProvider(*addonsProvidersRes.Payload(), "config-provider")
					plan := pkg.LookupProviderPlan(prov, "std")
					if plan == nil {
						t.Fatal("failed to find std plan for config-provider")
					}

					res := tmp.CreateAddon(ctx, cc, tests.ORGANISATION, tmp.AddonRequest{
						Name:       rName,
						Plan:       plan.ID,
						ProviderID: "config-provider",
						Region:     "par",
					})
					if res.HasError() {
						t.Fatalf("failed to create config provider addon: %s", res.Error())
					}
					addonID = res.Payload().ID
					realID = res.Payload().RealID

					// The fixture has to carry the same environment as the
					// config, otherwise the plan below legitimately wants to add
					// it and the step fails for the wrong reason.
					envRes := tmp.UpdateConfigProviderEnv(ctx, cc, tests.ORGANISATION, realID, tmp.EnvVars{
						{Name: "foo", Value: "this is foo"},
					})
					if envRes.HasError() {
						t.Fatalf("failed to set config provider env: %s", envRes.Error())
					}
				},
				Config:             providerBlock.Append(configProviderBlock).String(),
				ResourceName:       fullName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return realID, nil
				},
			},
			{
				// Re-apply the same config after import: the plan must be empty.
				// name is required and the env endpoint says nothing about the
				// add-on, so without reading it back from the add-on it stays
				// null in state and this step fails.
				Config:   providerBlock.Append(configProviderBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
