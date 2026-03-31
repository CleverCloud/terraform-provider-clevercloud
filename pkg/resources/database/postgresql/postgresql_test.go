package postgresql_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccPostgreSQL_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	rName := acctest.RandomWithPrefix("tf-test-pg")
	rNameEdited := rName + "-edit"
	rName2 := acctest.RandomWithPrefix("tf-test2-pg")
	fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_postgresql.%s", rName2)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "dev",
			"backup": true,
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^postgresql_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("dev")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("backup"), knownvalue.Bool(true)),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
			},
		}, {
			ResourceName: rName2,
			Config: providerBlock.Append(helper.NewRessource(
				"clevercloud_postgresql",
				rName2,
				helper.SetKeyValues(map[string]any{
					"name":    rName2,
					"region":  "par",
					"plan":    "xs_sml",
					"version": "17",
					"backup":  true,
				}))).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^postgresql_.*`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName2, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("plan"), knownvalue.StringExact("xs_sml")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("version"), knownvalue.StringExact("17")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("backup"), knownvalue.Bool(true)),
			},
		}, /*{
			ResourceName: rName3,
			Config: providerBlock.Append(helper.NewRessource(
				"clevercloud_postgresql",
				rName3,
				helper.SetKeyValues(map[string]any{
					"name":    rName2,
					"region":  "par",
					"plan":    "dev",
					"version": "10",
				}))).String(),
			ExpectError: regexp.MustCompile(`Error running pre-apply plan`),
		},*/
		},
	})
}

func TestAccPostgreSQL_RefreshDeleted(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-pg")
	//fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "dev",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock).String(),
		}, {
			ResourceName: rName,
			PreConfig: func() {
				tmp.DeleteAddon(context.Background(), cc, tests.ORGANISATION, rName)
			},
			RefreshState:       true,
			ExpectNonEmptyPlan: true,
		}},
	})
}

func TestAccPostgreSQL_migration(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	rName := acctest.RandomWithPrefix("tf-test-pg")
	fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xxs_sml",
		}))
	providerBlock2 := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	postgresqlBlock2 := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "m_med",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^postgresql_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("xxs_sml")),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
			},
		}, {
			ResourceName: rName,
			PreConfig: func() {
				// in order to test migrations, we must be sure the initial instance is started
				// there is not available information to know that yet
				time.Sleep(1 * time.Minute)
			},
			Config: providerBlock2.Append(postgresqlBlock2).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^postgresql_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-postgresql\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("m_med")),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
			},
		}},
	})
}

// TestAccPostgreSQL_EncryptionOnDevPlan reproduces issue #367
// When encryption=true is set on a 'dev' plan PostgreSQL, the provider should validate this
// and provide a clear error message since dev plans don't support encryption.
func TestAccPostgreSQL_EncryptionOnDevPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-pg-enc")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":       rName,
			"region":     "par",
			"plan":       "dev",
			"encryption": true,
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock).String(),
			ExpectError:  regexp.MustCompile(`Encryption not supported on shared plans`),
		}},
	})
}

// TestAccPostgreSQL_LocaleOnDevPlan tests that locale option is not allowed on dev plan
func TestAccPostgreSQL_LocaleOnDevPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-pg-locale")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "dev",
			"locale": "fr_FR",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock).String(),
			ExpectError:  regexp.MustCompile(`Locale not supported on shared plans`),
		}},
	})
}

// TestAccPostgreSQL_LocaleImport tests importing a PostgreSQL resource
// This validates that import works correctly and preserves the locale value
func TestAccPostgreSQL_LocaleImport(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	rName := acctest.RandomWithPrefix("tf-test-pg-locale-import")
	fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)

	configStr := helper.NewProvider("clevercloud").
		SetOrganisation(tests.ORGANISATION).
		Append(helper.NewRessource(
			"clevercloud_postgresql",
			rName,
			helper.SetKeyValues(map[string]any{
				"name":   rName,
				"region": "par",
				"plan":   "xs_sml",
				"backup": true,
			}))).String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: fullName,
			Config:       configStr,
		}, {
			ResourceName: fullName,
			ImportState:  true,
			//ImportStateId:     addon.RealID,
			ImportStateVerify: true,
		}},
	})
}

// TestAccPostgreSQL_LocaleCreate tests creating a PostgreSQL with a custom locale
// This validates that the fix allows using locale as a string (e.g., 'fr_FR')
func TestAccPostgreSQL_LocaleCreate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-pg-locale-create")
	fullName := fmt.Sprintf("clevercloud_postgresql.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Create with a custom locale on a dedicated plan
	postgresqlBlock := helper.NewRessource(
		"clevercloud_postgresql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs_sml", // dedicated plan that supports locale
			"locale": "fr_FR",  // Custom locale
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(postgresqlBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("locale"), knownvalue.StringExact("fr_FR")),
			},
		}},
	})
}
