package mysql_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

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

func TestAccMySQL_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-my")
	rNameEdited := rName + "-edit"
	rName2 := acctest.RandomWithPrefix("tf-test2-my")
	fullName := fmt.Sprintf("clevercloud_mysql.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_mysql.%s", rName2)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mysqlBlock := helper.NewRessource(
		"clevercloud_mysql",
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
			Config:       providerBlock.Append(mysqlBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^mysql_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-mysql\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("dev")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("backup"), knownvalue.Bool(true)),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(mysqlBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
			},
		}, {
			ResourceName: rName2,
			Config: providerBlock.Append(helper.NewRessource(
				"clevercloud_mysql",
				rName2,
				helper.SetKeyValues(map[string]any{
					"name":    rName2,
					"region":  "par",
					"plan":    "xs_sml",
					"version": "8.4",
					"backup":  true,
				}))).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^mysql_.*`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-mysql\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("password"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("plan"), knownvalue.StringExact("xs_sml")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("version"), knownvalue.StringExact("8.4")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("backup"), knownvalue.Bool(true)),
			},
		},
		},
	})
}

func TestAccMySQL_RefreshDeleted(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-my")
	//fullName := fmt.Sprintf("clevercloud_mysql.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mysqlBlock := helper.NewRessource(
		"clevercloud_mysql",
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
		Steps: []resource.TestStep{
			// create a database instance on first step
			{
				ResourceName: rName,
				Config:       providerBlock.Append(mysqlBlock).String(),
			},
			{
				ResourceName: rName,
				PreConfig: func() {
					// delete the database using an api call
					tmp.DeleteAddon(context.Background(), cc, tests.ORGANISATION, rName)
				},
				// refreshing state
				RefreshState: true,
				// plan should contain database re-creation
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMySQL_EncryptionOnDevPlan tests that encryption option is not allowed on dev plan
// Similar to PostgreSQL issue #367
func TestAccMySQL_EncryptionOnDevPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-my-enc")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mysqlBlock := helper.NewRessource(
		"clevercloud_mysql",
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
			Config:       providerBlock.Append(mysqlBlock).String(),
			ExpectError:  regexp.MustCompile(`Encryption not supported on shared plans`),
		}},
	})
}

// TestAccMySQL_SkipLogBinOnDevPlan tests that skip_log_bin option is not allowed on dev plan
func TestAccMySQL_SkipLogBinOnDevPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-my-skiplog")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mysqlBlock := helper.NewRessource(
		"clevercloud_mysql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":         rName,
			"region":       "par",
			"plan":         "dev",
			"skip_log_bin": true,
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(mysqlBlock).String(),
			ExpectError:  regexp.MustCompile(`Skip log bin not supported on shared plans`),
		}},
	})
}

// TestAccMySQL_DirectHostOnlyOnDevPlan tests that direct_host_only option is not allowed on dev plan
func TestAccMySQL_DirectHostOnlyOnDevPlan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-my-direct")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mysqlBlock := helper.NewRessource(
		"clevercloud_mysql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":             rName,
			"region":           "par",
			"plan":             "dev",
			"direct_host_only": true,
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(mysqlBlock).String(),
			ExpectError:  regexp.MustCompile(`Direct host only not supported on shared plans`),
		}},
	})
}

// TestAccMySQL_Import tests that importing a MySQL addon doesn't trigger
// a force-replace on the next plan.
// Reproduces GitHub issue #399: backup and skip_log_bin remain null in state
// after import because Read doesn't populate them when the API omits features.
// The null→default diff triggers RequiresReplace.
func TestAccMySQL_Import(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-my-import")
	fullName := fmt.Sprintf("clevercloud_mysql.%s", rName)

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
	mysqlBlock := helper.NewRessource(
		"clevercloud_mysql",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs_sml",
			"backup": true,
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create MySQL addon manually via API to simulate a pre-existing
				// resource, then import it into Terraform state.
				PreConfig: func() {
					addonsProvidersRes := tmp.GetAddonsProviders(ctx, cc)
					if addonsProvidersRes.HasError() {
						t.Fatalf("failed to get addon providers: %s", addonsProvidersRes.Error())
					}
					prov := pkg.LookupAddonProvider(*addonsProvidersRes.Payload(), "mysql-addon")
					plan := pkg.LookupProviderPlan(prov, "xs_sml")
					if plan == nil {
						t.Fatal("failed to find xs_sml plan for mysql-addon")
					}

					res := tmp.CreateAddon(ctx, cc, tests.ORGANISATION, tmp.AddonRequest{
						Name:       rName,
						Plan:       plan.ID,
						ProviderID: "mysql-addon",
						Region:     "par",
						Options:    map[string]string{"do-backup": "true"},
					})
					if res.HasError() {
						t.Fatalf("failed to create mysql addon: %s", res.Error())
					}
					addonID = res.Payload().ID
					realID = res.Payload().RealID
				},
				Config:       providerBlock.Append(mysqlBlock).String(),
				ResourceName: fullName,
				ImportState:  true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return realID, nil
				},
			},
			{
				// Re-apply the same config after import.
				// If backup/skip_log_bin are null in state (the bug), the schema
				// defaults will create a diff and RequiresReplace will trigger
				// a force-replace plan, causing this step to fail.
				Config:   providerBlock.Append(mysqlBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
