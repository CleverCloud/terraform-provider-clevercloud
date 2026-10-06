package redis_test

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

func TestAccRedis_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-redis")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_redis.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	materiakvBlock := helper.NewRessource("clevercloud_redis", rName, helper.SetKeyValues(map[string]any{
		"name":   rName,
		"region": "par",
		"plan":   "m_mono",
	}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^redis_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(materiakvBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^redis_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("port"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("token")),
			},
		}},
	})
}

// An imported add-on must plan empty. Read is the only import path, so every
// Computed attribute it forgets stays null and the schema default materialises
// on the next plan (#404). creation_date was exactly that case here.
func TestAccRedis_Import(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-redis-import")
	fullName := fmt.Sprintf("clevercloud_redis.%s", rName)

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
	redisBlock := helper.NewRessource("clevercloud_redis", rName, helper.SetKeyValues(map[string]any{
		"name":   rName,
		"region": "par",
		"plan":   "m_mono",
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
					prov := pkg.LookupAddonProvider(*providersRes.Payload(), "redis-addon")
					plan := pkg.LookupProviderPlan(prov, "m_mono")
					if plan == nil {
						t.Fatal("failed to find the m_mono plan for redis-addon")
					}

					res := tmp.CreateAddon(ctx, cc, tests.ORGANISATION, tmp.AddonRequest{
						Name:       rName,
						Plan:       plan.ID,
						ProviderID: "redis-addon",
						Region:     "par",
					})
					if res.HasError() {
						t.Fatalf("failed to create redis addon: %s", res.Error())
					}
					addonID = res.Payload().ID
					realID = res.Payload().RealID
				},
				Config:             providerBlock.Append(redisBlock).String(),
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
				Config:   providerBlock.Append(redisBlock).String(),
				PlanOnly: true,
			},
		},
	})
}
