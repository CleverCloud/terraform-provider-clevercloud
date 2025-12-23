package application_test

import (
	"context"
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
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// DepsResult holds both app and addon dependencies for verification
type DepsResult struct {
	Apps   []tmp.AppResponse
	Addons []tmp.AddonResponse
}

// fetchDependencies retrieves both app and addon dependencies from the API
func fetchDependencies(ctx context.Context, cc *client.Client, org, appID string) (*DepsResult, error) {
	result := &DepsResult{}

	appsRes := tmp.GetAppDependencies(ctx, cc, org, appID)
	if appsRes.HasError() {
		return nil, appsRes.Error()
	}
	result.Apps = *appsRes.Payload()

	addonsRes := tmp.GetAppLinkedAddons(ctx, cc, org, appID)
	if addonsRes.HasError() {
		return nil, addonsRes.Error()
	}
	result.Addons = *addonsRes.Payload()

	return result, nil
}

// TestAccDependencies_app_only tests app-to-app dependencies
func TestAccDependencies_app_only(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-deps")
	mainAppName := rName + "-main"
	depAppName := rName + "-dep"
	fullMainName := fmt.Sprintf("clevercloud_java_war.%s", mainAppName)
	fullDepName := fmt.Sprintf("clevercloud_java_war.%s", depAppName)
	cc := client.New(client.WithAutoOauthConfig())

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Dependency app (no dependencies itself)
	depAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		depAppName,
		helper.SetKeyValues(map[string]any{
			"name":               depAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}))

	// Main app depends on dep app
	mainAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		mainAppName,
		helper.SetKeyValues(map[string]any{
			"name":               mainAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"dependencies":       []string{fmt.Sprintf("${%s.id}", fullDepName)},
		}))

	config := providerBlock.Append(depAppBlock, mainAppBlock).String()

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             checkAppsDestroyed(ctx, cc),
		Steps: []resource.TestStep{{
			ResourceName: mainAppName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(1)),
				tests.NewCheckRemoteResource(fullMainName,
					func(ctx context.Context, id string) (*DepsResult, error) {
						return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
					},
					func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
						if len(deps.Apps) != 1 {
							return fmt.Errorf("expected 1 app dependency, got %d", len(deps.Apps))
						}
						if len(deps.Addons) != 0 {
							return fmt.Errorf("expected 0 addon dependencies, got %d", len(deps.Addons))
						}
						return nil
					}),
			},
		}},
	})
}

// TestAccDependencies_addon_only tests addon dependencies
func TestAccDependencies_addon_only(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-deps")
	mainAppName := rName + "-main"
	pgName := rName + "-pg"
	fullMainName := fmt.Sprintf("clevercloud_java_war.%s", mainAppName)
	fullPgName := fmt.Sprintf("clevercloud_postgresql.%s", pgName)
	cc := client.New(client.WithAutoOauthConfig())

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// PostgreSQL addon
	pgBlock := helper.NewRessource(
		"clevercloud_postgresql",
		pgName,
		helper.SetKeyValues(map[string]any{
			"name":   pgName,
			"region": "par",
			"plan":   "dev",
		}))

	// Main app depends on PostgreSQL
	mainAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		mainAppName,
		helper.SetKeyValues(map[string]any{
			"name":               mainAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"dependencies":       []string{fmt.Sprintf("${%s.id}", fullPgName)},
		}))

	config := providerBlock.Append(pgBlock, mainAppBlock).String()

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             checkResourcesDestroyed(ctx, cc),
		Steps: []resource.TestStep{{
			ResourceName: mainAppName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(1)),
				tests.NewCheckRemoteResource(fullMainName,
					func(ctx context.Context, id string) (*DepsResult, error) {
						return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
					},
					func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
						if len(deps.Apps) != 0 {
							return fmt.Errorf("expected 0 app dependencies, got %d", len(deps.Apps))
						}
						if len(deps.Addons) != 1 {
							return fmt.Errorf("expected 1 addon dependency, got %d", len(deps.Addons))
						}
						return nil
					}),
			},
		}},
	})
}

// TestAccDependencies_mixed tests both app and addon dependencies together
func TestAccDependencies_mixed(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-deps")
	mainAppName := rName + "-main"
	depAppName := rName + "-dep"
	pgName := rName + "-pg"
	fullMainName := fmt.Sprintf("clevercloud_java_war.%s", mainAppName)
	fullDepName := fmt.Sprintf("clevercloud_java_war.%s", depAppName)
	fullPgName := fmt.Sprintf("clevercloud_postgresql.%s", pgName)
	cc := client.New(client.WithAutoOauthConfig())

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Dependency app
	depAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		depAppName,
		helper.SetKeyValues(map[string]any{
			"name":               depAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}))

	// PostgreSQL addon
	pgBlock := helper.NewRessource(
		"clevercloud_postgresql",
		pgName,
		helper.SetKeyValues(map[string]any{
			"name":   pgName,
			"region": "par",
			"plan":   "dev",
		}))

	// Main app depends on both app and addon
	mainAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		mainAppName,
		helper.SetKeyValues(map[string]any{
			"name":               mainAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"dependencies": []string{
				fmt.Sprintf("${%s.id}", fullDepName),
				fmt.Sprintf("${%s.id}", fullPgName),
			},
		}))

	config := providerBlock.Append(depAppBlock, pgBlock, mainAppBlock).String()

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             checkResourcesDestroyed(ctx, cc),
		Steps: []resource.TestStep{{
			ResourceName: mainAppName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(2)),
				tests.NewCheckRemoteResource(fullMainName,
					func(ctx context.Context, id string) (*DepsResult, error) {
						return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
					},
					func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
						if len(deps.Apps) != 1 {
							return fmt.Errorf("expected 1 app dependency, got %d", len(deps.Apps))
						}
						if len(deps.Addons) != 1 {
							return fmt.Errorf("expected 1 addon dependency, got %d", len(deps.Addons))
						}
						return nil
					}),
			},
		}},
	})
}

// TestAccDependencies_sync tests adding and removing dependencies across multiple steps
func TestAccDependencies_sync(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-deps")
	mainAppName := rName + "-main"
	depAppName := rName + "-dep"
	pgName := rName + "-pg"
	fullMainName := fmt.Sprintf("clevercloud_java_war.%s", mainAppName)
	fullDepName := fmt.Sprintf("clevercloud_java_war.%s", depAppName)
	fullPgName := fmt.Sprintf("clevercloud_postgresql.%s", pgName)
	cc := client.New(client.WithAutoOauthConfig())

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Dependency app (always present)
	depAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		depAppName,
		helper.SetKeyValues(map[string]any{
			"name":               depAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}))

	// PostgreSQL addon (always present)
	pgBlock := helper.NewRessource(
		"clevercloud_postgresql",
		pgName,
		helper.SetKeyValues(map[string]any{
			"name":   pgName,
			"region": "par",
			"plan":   "dev",
		}))

	// Main app block - will be mutated between steps
	mainAppBlock := helper.NewRessource(
		"clevercloud_java_war",
		mainAppName,
		helper.SetKeyValues(map[string]any{
			"name":               mainAppName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             checkResourcesDestroyed(ctx, cc),
		Steps: []resource.TestStep{
			// Step 1: No dependencies
			{
				ResourceName: mainAppName,
				Config:       providerBlock.Append(depAppBlock, pgBlock, mainAppBlock).String(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
					tests.NewCheckRemoteResource(fullMainName,
						func(ctx context.Context, id string) (*DepsResult, error) {
							return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
						},
						func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
							if len(deps.Apps) != 0 {
								return fmt.Errorf("step1: expected 0 app dependencies, got %d", len(deps.Apps))
							}
							if len(deps.Addons) != 0 {
								return fmt.Errorf("step1: expected 0 addon dependencies, got %d", len(deps.Addons))
							}
							return nil
						}),
				},
			},
			// Step 2: Add app dependency
			{
				ResourceName: mainAppName,
				Config: providerBlock.Append(depAppBlock, pgBlock,
					mainAppBlock.SetOneValue("dependencies", []string{fmt.Sprintf("${%s.id}", fullDepName)}),
				).String(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(1)),
					tests.NewCheckRemoteResource(fullMainName,
						func(ctx context.Context, id string) (*DepsResult, error) {
							return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
						},
						func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
							if len(deps.Apps) != 1 {
								return fmt.Errorf("step2: expected 1 app dependency, got %d", len(deps.Apps))
							}
							if len(deps.Addons) != 0 {
								return fmt.Errorf("step2: expected 0 addon dependencies, got %d", len(deps.Addons))
							}
							return nil
						}),
				},
			},
			// Step 3: Add addon dependency (now have both)
			{
				ResourceName: mainAppName,
				Config: providerBlock.Append(depAppBlock, pgBlock,
					mainAppBlock.SetOneValue("dependencies", []string{
						fmt.Sprintf("${%s.id}", fullDepName),
						fmt.Sprintf("${%s.id}", fullPgName),
					}),
				).String(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(2)),
					tests.NewCheckRemoteResource(fullMainName,
						func(ctx context.Context, id string) (*DepsResult, error) {
							return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
						},
						func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
							if len(deps.Apps) != 1 {
								return fmt.Errorf("step3: expected 1 app dependency, got %d", len(deps.Apps))
							}
							if len(deps.Addons) != 1 {
								return fmt.Errorf("step3: expected 1 addon dependency, got %d", len(deps.Addons))
							}
							return nil
						}),
				},
			},
			// Step 4: Remove app dependency (keep only addon)
			{
				ResourceName: mainAppName,
				Config: providerBlock.Append(depAppBlock, pgBlock,
					mainAppBlock.SetOneValue("dependencies", []string{fmt.Sprintf("${%s.id}", fullPgName)}),
				).String(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(1)),
					tests.NewCheckRemoteResource(fullMainName,
						func(ctx context.Context, id string) (*DepsResult, error) {
							return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
						},
						func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
							if len(deps.Apps) != 0 {
								return fmt.Errorf("step4: expected 0 app dependencies, got %d", len(deps.Apps))
							}
							if len(deps.Addons) != 1 {
								return fmt.Errorf("step4: expected 1 addon dependency, got %d", len(deps.Addons))
							}
							return nil
						}),
				},
			},
			// Step 5: Remove all dependencies
			{
				ResourceName: mainAppName,
				Config: providerBlock.Append(depAppBlock, pgBlock,
					mainAppBlock.SetOneValue("dependencies", []string{}),
				).String(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullMainName, tfjsonpath.New("dependencies"), knownvalue.SetSizeExact(0)),
					tests.NewCheckRemoteResource(fullMainName,
						func(ctx context.Context, id string) (*DepsResult, error) {
							return fetchDependencies(ctx, cc, tests.ORGANISATION, id)
						},
						func(ctx context.Context, id string, state *tfjson.State, deps *DepsResult) error {
							if len(deps.Apps) != 0 {
								return fmt.Errorf("step5: expected 0 app dependencies, got %d", len(deps.Apps))
							}
							if len(deps.Addons) != 0 {
								return fmt.Errorf("step5: expected 0 addon dependencies, got %d", len(deps.Addons))
							}
							return nil
						}),
				},
			},
		},
	})
}

// checkAppsDestroyed verifies that all apps are deleted after test
func checkAppsDestroyed(ctx context.Context, cc *client.Client) func(*terraform.State) error {
	return func(state *terraform.State) error {
		for _, res := range state.RootModule().Resources {
			if res.Type != "clevercloud_java_war" {
				continue
			}
			appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, res.Primary.ID)
			if appRes.IsNotFoundError() {
				continue
			}
			if appRes.HasError() {
				return fmt.Errorf("unexpected error: %s", appRes.Error().Error())
			}
			if appRes.Payload().State == "TO_DELETE" {
				continue
			}
			return fmt.Errorf("expect resource '%s' to be deleted, state: '%s'", res.Primary.ID, appRes.Payload().State)
		}
		return nil
	}
}

// checkResourcesDestroyed verifies that all apps and addons are deleted after test
func checkResourcesDestroyed(ctx context.Context, cc *client.Client) func(*terraform.State) error {
	return func(state *terraform.State) error {
		for _, res := range state.RootModule().Resources {
			switch res.Type {
			case "clevercloud_java_war":
				appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, res.Primary.ID)
				if appRes.IsNotFoundError() {
					continue
				}
				if appRes.HasError() {
					return fmt.Errorf("unexpected error: %s", appRes.Error().Error())
				}
				if appRes.Payload().State == "TO_DELETE" {
					continue
				}
				return fmt.Errorf("expect app '%s' to be deleted, state: '%s'", res.Primary.ID, appRes.Payload().State)

			case "clevercloud_postgresql":
				addonID, err := tmp.RealIDToAddonID(ctx, cc, tests.ORGANISATION, res.Primary.ID)
				if err != nil {
					// Not found is OK
					continue
				}
				pgRes := tmp.GetPostgreSQL(ctx, cc, addonID)
				if pgRes.IsNotFoundError() {
					continue
				}
				if pgRes.HasError() {
					return fmt.Errorf("unexpected error: %s", pgRes.Error().Error())
				}
				if pgRes.Payload().Status == "TO_DELETE" {
					continue
				}
				return fmt.Errorf("expect addon '%s' to be deleted", res.Primary.ID)
			}
		}
		return nil
	}
}
