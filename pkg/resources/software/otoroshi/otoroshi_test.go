package otoroshi_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"
	"time"

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

func TestAccOtoroshi_basic(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-otoroshi")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_otoroshi.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	otoroshiBlock := helper.NewRessource(
		"clevercloud_otoroshi",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "otoroshi_" + rName,
			Config:       providerBlock.Append(otoroshiBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^otoroshi_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-api-otoroshi\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_login"), knownvalue.StringExact("cc-account-admin")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_password"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-ui-otoroshi\.services\.clever-cloud\.com$`))),
			},
		}, {
			ResourceName: "otoroshi_" + rName,
			Config:       providerBlock.Append(otoroshiBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^otoroshi_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_client_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("api_url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-api-otoroshi\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_login"), knownvalue.StringExact("cc-account-admin")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("initial_admin_password"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("url"), knownvalue.StringRegexp(regexp.MustCompile(`^https://.*-ui-otoroshi\.services\.clever-cloud\.com$`))),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccOtoroshi_networkgroup(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-otoroshi")
	ngName := acctest.RandomWithPrefix("tf-test-otoroshi-ng")
	fullName := fmt.Sprintf("clevercloud_otoroshi.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	ngBlock := helper.NewRessource(
		"clevercloud_networkgroup",
		ngName,
		helper.SetKeyValues(map[string]any{
			"name":        ngName,
			"description": "for otoroshi ng tests",
			"tags":        []string{"otoroshi"},
		}),
	)

	otoroshiBlock := helper.NewRessource(
		"clevercloud_otoroshi",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"networkgroups": []map[string]string{{
				"networkgroup_id": fmt.Sprintf("${clevercloud_networkgroup.%s.id}", ngName),
				"fqdn":            "myotoroshi",
			}}}),
	)

	otoroshiBlock2 := helper.NewRessource(
		"clevercloud_otoroshi",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":          rName,
			"region":        "par",
			"networkgroups": nil,
		}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: "otoroshi_" + rName,
			Config:       providerBlock.Append(ngBlock, otoroshiBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^otoroshi_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("networkgroups"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"networkgroup_id": knownvalue.StringRegexp(regexp.MustCompile("^ng_.*$")),
						"fqdn":            knownvalue.StringExact("myotoroshi"),
					}),
				})),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					ngResource := s.RootModule().Resources["clevercloud_networkgroup."+ngName]
					otoroshiResource := s.RootModule().Resources[fullName]
					ngID := ngResource.Primary.ID

					otoroshiRes := tmp.GetOtoroshi(ctx, cc, otoroshiResource.Primary.ID)
					if otoroshiRes.HasError() {
						return fmt.Errorf("failed to get otoroshi: %w", otoroshiRes.Error())
					}
					entrypoint := otoroshiRes.Payload().Resources.Entrypoint

					membersRes := tmp.ListMembers(ctx, cc, tests.ORGANISATION, ngID)
					if membersRes.HasError() {
						return fmt.Errorf("failed to list members: %w", membersRes.Error())
					}
					members := *membersRes.Payload()

					if len(members) != 1 {
						return fmt.Errorf("expect 1 member, got: %d", len(members))
					}
					member := members[0]

					if member.ID != entrypoint {
						return fmt.Errorf("expect member to be the underlying app %s, got: %s", entrypoint, member.ID)
					}
					return nil
				},
			),
		}, {
			ResourceName: "otoroshi_" + rName,
			Config:       providerBlock.Append(ngBlock, otoroshiBlock2).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^otoroshi_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("networkgroups"), knownvalue.Null()),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					time.Sleep(5 * time.Second) // NG API is asynchronous
					ngResource := s.RootModule().Resources["clevercloud_networkgroup."+ngName]
					ngID := ngResource.Primary.ID

					membersRes := tmp.ListMembers(ctx, cc, tests.ORGANISATION, ngID)
					if membersRes.HasError() {
						return fmt.Errorf("failed to list members: %w", membersRes.Error())
					}
					members := *membersRes.Payload()

					if len(members) != 0 {
						return fmt.Errorf("expect 0 member, got: %+v", members)
					}

					return nil
				},
			),
		}},
	})
}
