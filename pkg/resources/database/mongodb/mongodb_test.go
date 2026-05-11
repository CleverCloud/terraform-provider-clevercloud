package mongodb_test

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

func TestAccMongoDB_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-mg")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_mongodb.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mongodbBlock := helper.NewRessource("clevercloud_mongodb", rName, helper.SetKeyValues(map[string]any{"name": rName, "plan": "xs_med", "region": "par"}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(mongodbBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^mongodb_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-mongodb\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(mongodbBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^mongodb_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*-mongodb\.services\.clever-cloud\.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("user"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("password")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("database"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
			},
		}},
	})
}

func TestAccMongoDB_RefreshDeleted(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-mg")
	//fullName := fmt.Sprintf("clevercloud_mongodb.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mongodbBlock2 := helper.NewRessource("clevercloud_mongodb", rName, helper.SetKeyValues(map[string]any{"name": rName, "plan": "xs_med", "region": "par"}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			// create a database instance on first step
			{
				ResourceName: rName,
				Config:       providerBlock.Append(mongodbBlock2).String(),
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

// TestAccMongoDB_Import tests that importing a MongoDB addon doesn't trigger
// a force-replace on the next plan.
// Same class of bug as MySQL issue #399: feature attributes remain null in state
// after import because Read doesn't populate them when the API omits features.
func TestAccMongoDB_Import(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-mg-import")
	fullName := fmt.Sprintf("clevercloud_mongodb.%s", rName)

	var addonID string
	var realID string

	t.Cleanup(func() {
		if addonID != "" {
			tmp.DeleteAddon(context.Background(), cc, tests.ORGANISATION, addonID)
		}
	})

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	mgBlock := helper.NewRessource(
		"clevercloud_mongodb",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":   rName,
			"region": "par",
			"plan":   "xs_med",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					addonsProvidersRes := tmp.GetAddonsProviders(ctx, cc)
					if addonsProvidersRes.HasError() {
						t.Fatalf("failed to get addon providers: %s", addonsProvidersRes.Error())
					}
					prov := pkg.LookupAddonProvider(*addonsProvidersRes.Payload(), "mongodb-addon")
					plan := pkg.LookupProviderPlan(prov, "xs_med")
					if plan == nil {
						t.Fatal("failed to find xs_med plan for mongodb-addon")
					}

					res := tmp.CreateAddon(ctx, cc, tests.ORGANISATION, tmp.AddonRequest{
						Name:       rName,
						Plan:       plan.ID,
						ProviderID: "mongodb-addon",
						Region:     "par",
						Options:    map[string]string{},
					})
					if res.HasError() {
						t.Fatalf("failed to create mongodb addon: %s", res.Error())
					}
					addonID = res.Payload().ID
					realID = res.Payload().RealID
				},
				Config:             providerBlock.Append(mgBlock).String(),
				ResourceName:       fullName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return realID, nil
				},
			},
			{
				Config:   providerBlock.Append(mgBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
