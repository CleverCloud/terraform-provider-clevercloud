package metabase_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

func TestAccMetabase_basic(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-mb")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_metabase.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).
		SetKeyValues(map[string]any{"default_tags": []string{"managed"}})
	metabaseBlock := helper.NewRessource(
		"clevercloud_metabase",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par", "tags": []string{"foo", "bar"}}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(metabaseBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^metabase_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("foo"),
				})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags_all"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("foo"),
					knownvalue.StringExact("managed"),
				})),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(metabaseBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^metabase_.*`))),
			},
		}},
	})
}
