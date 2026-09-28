package addon_test

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
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccAddon_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	rName := acctest.RandomWithPrefix("tf-test-mp")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_addon.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	addonBlock := helper.NewRessource(
		"clevercloud_addon",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":                 rName,
			"region":               "par",
			"plan":                 "clever_solo",
			"third_party_provider": "mailpace",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^addon_.*`))),
				// TODO test env var existance
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(addonBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^addon_.*`))),
			},
		}},
	})
}

// The API spells some plan slugs in uppercase — cellar answers "S", jenkins
// "XS", "S", "M". A lowercase-only validator on plan used to reject them, which
// made those add-ons impossible to import: Read copies the slug the API returns,
// and the plan then failed validation on the provider's own value.
func TestAccAddon_UppercasePlan(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-addon-upper")
	fullName := fmt.Sprintf("clevercloud_addon.%s", rName)

	var addonID string

	t.Cleanup(func() {
		if addonID != "" {
			tmp.DeleteAddon(context.Background(), cc, tests.ORGANISATION, addonID)
		}
	})

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	addonBlock := helper.NewRessource(
		"clevercloud_addon",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":                 rName,
			"region":               "par",
			"plan":                 "S", // as the provider spells it
			"third_party_provider": "cellar-addon",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       providerBlock.Append(addonBlock).String(),
				ResourceName: rName,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("S")),
				},
			},
			{
				// The uppercase slug must survive a refresh, not drift.
				Config:   providerBlock.Append(addonBlock).String(),
				PlanOnly: true,
			},
		},
	})
}

// Lowercase stays accepted: LookupProviderPlan matches case-insensitively, so
// both spellings name the same plan and neither should drift.
func TestAccAddon_LowercasePlanDoesNotDrift(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-addon-lower")
	fullName := fmt.Sprintf("clevercloud_addon.%s", rName)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	addonBlock := helper.NewRessource(
		"clevercloud_addon",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":                 rName,
			"region":               "par",
			"plan":                 "s", // the provider answers "S"
			"third_party_provider": "cellar-addon",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       providerBlock.Append(addonBlock).String(),
				ResourceName: rName,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("s")),
				},
			},
			{
				Config:   providerBlock.Append(addonBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
