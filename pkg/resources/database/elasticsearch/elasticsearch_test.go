package elasticsearch_test

import (
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
		CheckDestroy:             tests.CheckDestroy(ctx),
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

// Issue #338: Test that version handling works correctly
// The version field accepts only major version numbers (e.g., "8").
// The API may return "8" or "8.19.9" but we always extract and store the major version.
func TestAccElasticsearch_VersionDrift(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-es")
	fullName := fmt.Sprintf("clevercloud_elasticsearch.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	elasticsearchBlock := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":    rName,
			"region":  "par",
			"plan":    "xs",
			"version": "8",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(elasticsearchBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[fullName]
					if !ok {
						return fmt.Errorf("resource %s not found", fullName)
					}
					version := rs.Primary.Attributes["version"]

					if version != "8" {
						return fmt.Errorf("expected version to be '8', got: %s", version)
					}
					return nil
				},
				resource.TestCheckResourceAttr(fullName, "version", "8"),
			),
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(elasticsearchBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[fullName]
					if !ok {
						return fmt.Errorf("resource %s not found", fullName)
					}
					version := rs.Primary.Attributes["version"]

					if version != "8" {
						return fmt.Errorf("expected version to remain '8', got: %s", version)
					}
					return nil
				},
				resource.TestCheckResourceAttr(fullName, "version", "8"),
			),
		}, {
			ResourceName:       rName,
			Config:             providerBlock.Append(elasticsearchBlock).String(),
			PlanOnly:           true,
			ExpectNonEmptyPlan: false,
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(elasticsearchBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[fullName]
					if !ok {
						return fmt.Errorf("resource %s not found", fullName)
					}
					version := rs.Primary.Attributes["version"]
					if version != "8" {
						return fmt.Errorf("expected final version to be '8', got: %s", version)
					}
					return nil
				},
				resource.TestCheckResourceAttr(fullName, "version", "8"),
				resource.TestCheckResourceAttrSet(fullName, "host"),
				resource.TestCheckResourceAttrSet(fullName, "user"),
				resource.TestCheckResourceAttrSet(fullName, "password"),
			),
		}},
	})
}

func TestAccElasticsearch_InvalidVersion(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-es")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Try to create with an invalid version (should fail validation)
	elasticsearchBlock := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":    rName,
			"region":  "par",
			"plan":    "xs",
			"version": "999", // Invalid version that doesn't exist
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(elasticsearchBlock).String(),
			ExpectError:  regexp.MustCompile("version '999' is not available"),
		}},
	})
}

func TestAccElasticsearch_InvalidVersionFormat(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-es")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Try to create with a full semver instead of major version (should fail format validation)
	elasticsearchBlock := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":    rName,
			"region":  "par",
			"plan":    "xs",
			"version": "8.19.7", // Invalid: should be just "8"
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(elasticsearchBlock).String(),
			ExpectError:  regexp.MustCompile("version must be a major version number"),
		}},
	})
}

// Issue #357: Test that changing immutable features (kibana, apm, encryption) triggers replacement
// When an Elasticsearch addon is created with kibana=false, changing it to kibana=true should
// force the resource to be replaced since there is no API to enable/disable these features after creation.
func TestAccElasticsearch_KibanaRequiresReplace(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-es")
	fullName := fmt.Sprintf("clevercloud_elasticsearch.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Step 1: Create ES without Kibana
	elasticsearchBlock := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs",
			"kibana": false,
		}))

	// Step 2: Enable Kibana (should trigger replace)
	elasticsearchBlockWithKibana := helper.NewRessource(
		"clevercloud_elasticsearch",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs",
			"kibana": true,
		}))

	var firstID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ResourceName: rName,
				Config:       providerBlock.Append(elasticsearchBlock).String(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(fullName, "kibana", "false"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[fullName]
						if !ok {
							return fmt.Errorf("resource %s not found", fullName)
						}
						firstID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				ResourceName: rName,
				Config:       providerBlock.Append(elasticsearchBlockWithKibana).String(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(fullName, "kibana", "true"),
					// Verify that the resource ID has changed (indicating replacement)
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[fullName]
						if !ok {
							return fmt.Errorf("resource %s not found", fullName)
						}
						if rs.Primary.ID == firstID {
							return fmt.Errorf("expected resource to be replaced (different ID), but ID remained the same: %s", firstID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccElasticsearch_RefreshDeleted(t *testing.T) {
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-es")
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
		CheckDestroy:             tests.CheckDestroy(ctx),
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
