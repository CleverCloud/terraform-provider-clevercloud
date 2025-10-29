package elasticsearch_test

import (
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

func TestAccElasticsearch_basic(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-es")
	//rNameEdited := rName + "-edit"
	//rName2 := acctest.RandomWithPrefix("tf-test2-es")
	fullName := fmt.Sprintf("clevercloud_elasticsearch.%s", rName)
	//fullName2 := fmt.Sprintf("clevercloud_elasticsearch.%s", rName2)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	elasticsearchBlock := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs",
		}))

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

				res := tmp.GetElasticsearch(ctx, cc, addonId)
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
			Config:       providerBlock.Append(elasticsearchBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectIdentityValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^elasticsearch_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("xs")),
			},
		}, /*{ // TODO: update
			ResourceName: rName,
			Config:       providerBlock.Append(elasticsearchBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
			},
		}, {
			ResourceName: rName2,
			Config: providerBlock.Append(helper.NewRessource(
				"clevercloud_elasticsearch",
				rName2,
				helper.SetKeyValues(map[string]any{
					"name":    rName2,
					"region":  "par",
					"plan":    "s",
					"version": "8",
					"backup":  true,
					"kibana":  true,
					"apm":     false,
				}))).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectIdentityValue(fullName2, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^elasticsearch_.*`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName2, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("plan"), knownvalue.StringExact("s")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("version"), knownvalue.StringExact("8")),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("kibana"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName2, tfjsonpath.New("apm"), knownvalue.Bool(false)),
			},
		}*/},
	})
}

func TestAccElasticsearch_RefreshDeleted(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-es")
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	elasticsearchBlock := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs",
		}))

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

				res := tmp.GetElasticsearch(ctx, cc, addonId)
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
			Config:       providerBlock.Append(elasticsearchBlock).String(),
		}, {
			ResourceName: rName,
			PreConfig: func() {
				// delete the elasticsearch using an api call
				tmp.DeleteAddon(ctx, cc, tests.ORGANISATION, rName)
			},
			// refreshing state
			RefreshState: true,
			// plan should contain elasticsearch re-creation
			ExpectNonEmptyPlan: true,
		}},
	})
}
