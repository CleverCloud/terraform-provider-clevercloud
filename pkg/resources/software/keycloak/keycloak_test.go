package keycloak_test

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"

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
	"go.clever-cloud.dev/sdk"
)

func TestAccKeycloak_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-kc")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_keycloak.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	materiakvBlock := helper.NewRessource(
		"clevercloud_keycloak",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetAddon(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect resource '%s' to be deleted: %+v", resource.Primary.ID, res.Payload())
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^keycloak_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*clever-cloud.com/admin$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("admin_username"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("admin_password"), knownvalue.NotNull()),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^keycloak_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*clever-cloud.com/admin$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("admin_username"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("admin_password"), knownvalue.NotNull()),
			},
		}},
	})
}

func TestAccKeycloak_invalidVersion(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tf-test-kc")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	keycloakBlock := helper.NewRessource(
		"clevercloud_keycloak",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":    rName,
			"region":  "par",
			"version": "99.99.99", // Version that doesn't exist
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(keycloakBlock).String(),
			ExpectError:  regexp.MustCompile("unavailable version"),
		}},
	})
}

func TestAccKeycloak_versionUpgrade(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip("no flag for running acceptance tests")
	}

	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-kc")
	fullName := fmt.Sprintf("clevercloud_keycloak.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	// Fetch available versions from API
	keycloakSDK := sdk.NewSDK(sdk.WithClient(cc))
	infosRes := keycloakSDK.V4().AddonProviders().Keycloak().Getkeycloakproviderinformation(ctx)
	if infosRes.HasError() {
		t.Fatalf("failed to get Keycloak provider information: %s", infosRes.Error().Error())
	}
	infos := infosRes.Payload()

	// Get sorted list of versions
	versions := make([]string, 0, len(infos.Dedicated))
	for version := range infos.Dedicated {
		versions = append(versions, version)
	}

	if len(versions) < 2 {
		t.Skip("need at least 2 versions to test upgrade")
	}

	firstVersion := versions[0]
	lastVersion := versions[len(versions)-1]

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	keycloakBlock := helper.NewRessource(
		"clevercloud_keycloak",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":    rName,
			"region":  "par",
			"version": firstVersion,
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck: func() {
			tests.ExpectOrganisation(t)
		},
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetAddon(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpected error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect resource '%s' to be deleted: %+v", resource.Primary.ID, res.Payload())
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(keycloakBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.StringExact(firstVersion)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^keycloak_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*clever-cloud.com/admin$`))),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(keycloakBlock.SetOneValue("version", lastVersion)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.StringExact(lastVersion)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^keycloak_.*`))),
			},
		}},
	})
}
