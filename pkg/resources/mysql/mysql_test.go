package mysql_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
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
)

func TestAccMySQL_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-my")
	rNameEdited := rName + "-edit"
	rName2 := acctest.RandomWithPrefix("tf-test2-my")
	fullName := fmt.Sprintf("clevercloud_mysql.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_mysql.%s", rName2)
	cc := client.New(client.WithAutoOauthConfig())
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
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				addonId, err := tmp.RealIDToAddonID(context.Background(), cc, tests.ORGANISATION, resource.Primary.ID)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						continue
					}
					return fmt.Errorf("failed to get addon ID: %s", err.Error())
				}

				res := tmp.GetMySQL(context.Background(), cc, addonId)
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
	rName := acctest.RandomWithPrefix("tf-test-my")
	//fullName := fmt.Sprintf("clevercloud_mysql.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
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
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				addonId, err := tmp.RealIDToAddonID(context.Background(), cc, tests.ORGANISATION, resource.Primary.ID)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						continue
					}
					return fmt.Errorf("failed to get addon ID: %s", err.Error())
				}

				res := tmp.GetMySQL(context.Background(), cc, addonId)
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
