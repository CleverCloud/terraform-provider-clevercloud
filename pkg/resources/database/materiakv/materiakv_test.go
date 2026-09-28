package materiakv_test

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

func TestAccMateriaKV_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-kv")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_materia_kv.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	materiakvBlock := helper.NewRessource("clevercloud_materia_kv", rName, helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^kv_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^kv_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}},
	})
}

func TestAccMateriaKV_Import(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-kv-import")
	fullName := fmt.Sprintf("clevercloud_materia_kv.%s", rName)

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
	materiakvBlock := helper.NewRessource(
		"clevercloud_materia_kv",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create the add-on through the API to simulate a pre-existing
				// resource, then import it into Terraform state.
				PreConfig: func() {
					addonsProvidersRes := tmp.GetAddonsProviders(ctx, cc)
					if addonsProvidersRes.HasError() {
						t.Fatalf("failed to get addon providers: %s", addonsProvidersRes.Error())
					}
					prov := pkg.LookupAddonProvider(*addonsProvidersRes.Payload(), "kv")
					plan := pkg.LookupProviderPlan(prov, "base")
					if plan == nil {
						t.Fatal("failed to find base plan for kv")
					}

					res := tmp.CreateAddon(ctx, cc, tests.ORGANISATION, tmp.AddonRequest{
						Name:       rName,
						Plan:       plan.ID,
						ProviderID: "kv",
						Region:     "par",
					})
					if res.HasError() {
						t.Fatalf("failed to create materia kv addon: %s", res.Error())
					}
					addonID = res.Payload().ID
					realID = res.Payload().RealID
				},
				Config:             providerBlock.Append(materiakvBlock).String(),
				ResourceName:       fullName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return realID, nil
				},
			},
			{
				// Re-apply the same config after import: the plan must be empty.
				// name is required and the Materia payload does not carry it, so
				// without reading it back from the add-on it stays null in state
				// and this step fails.
				Config:   providerBlock.Append(materiakvBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
