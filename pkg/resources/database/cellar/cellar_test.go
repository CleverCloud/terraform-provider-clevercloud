package cellar_test

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

func TestAccCellar_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-cellar")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_cellar.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	cellarBlock := helper.NewRessource(
		"clevercloud_cellar",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "cellar_" + rName,
			Config:       providerBlock.Append(cellarBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^cellar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Z0-9]{20}$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
			},
		}, {
			ResourceName: "cellar_" + rName,
			Config:       providerBlock.Append(cellarBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^cellar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Z0-9]{20}$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

// An imported add-on must plan empty. Read is the only import path, so every
// Computed attribute it forgets stays null and the schema default materialises
// on the next plan (#404).
func TestAccCellar_Import(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-cellar-import")
	fullName := fmt.Sprintf("clevercloud_cellar.%s", rName)

	var addonID string
	var realID string

	// Clean up the add-on created through the API, in case the import step fails
	// before Terraform takes ownership of it.
	t.Cleanup(func() {
		if addonID != "" {
			tmp.DeleteAddon(context.Background(), cc, tests.ORGANISATION, addonID)
		}
	})

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	cellarBlock := helper.NewRessource("clevercloud_cellar", rName, helper.SetKeyValues(map[string]any{
		"name":   rName,
		"region": "par",
	}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create the add-on through the API to stand in for a
				// pre-existing resource, then import it.
				PreConfig: func() {
					providersRes := tmp.GetAddonsProviders(ctx, cc)
					if providersRes.HasError() {
						t.Fatalf("failed to get addon providers: %s", providersRes.Error())
					}
					prov := pkg.LookupAddonProvider(*providersRes.Payload(), "cellar-addon")
					if len(prov.Plans) == 0 {
						t.Fatal("cellar-addon offers no plan")
					}

					res := tmp.CreateAddon(ctx, cc, tests.ORGANISATION, tmp.AddonRequest{
						Name:       rName,
						Plan:       prov.Plans[0].ID,
						ProviderID: "cellar-addon",
						Region:     "par",
					})
					if res.HasError() {
						t.Fatalf("failed to create cellar addon: %s", res.Error())
					}
					addonID = res.Payload().ID
					realID = res.Payload().RealID
				},
				Config:             providerBlock.Append(cellarBlock).String(),
				ResourceName:       fullName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return realID, nil
				},
			},
			{
				// The same config, planned against the imported state: an
				// attribute Read left null shows up here as a diff.
				Config:   providerBlock.Append(cellarBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
