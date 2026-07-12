package cellar_test

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
			"tags":   []string{"foo", "bar"},
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
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("foo"),
				})),
			},
		}, {
			ResourceName: "cellar_" + rName,
			Config: providerBlock.Append(
				cellarBlock.SetOneValue("name", rNameEdited).SetOneValue("tags", []string{"bar", "baz"}),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^cellar_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*\.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_id"), knownvalue.StringRegexp(regexp.MustCompile(`^[A-Z0-9]{20}$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("key_secret"), knownvalue.StringRegexp(regexp.MustCompile(`^[a-zA-Z0-9]+$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("baz"),
				})),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

// TestAccCellar_DefaultTags verifies that provider-level default_tags are merged into
// tags_all (without polluting the resource-level tags attribute), and that changing
// default_tags propagates to an existing resource via an in-place update.
func TestAccCellar_DefaultTags(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-cellar-tags")
	fullName := fmt.Sprintf("clevercloud_cellar.%s", rName)

	cellarBlock := helper.NewRessource(
		"clevercloud_cellar",
		rName,
		helper.SetKeyValues(map[string]any{
			"name": rName,
			"tags": []string{"foo", "bar"},
		}))

	// config builds the provider block with the given default_tags plus the cellar.
	config := func(defaultTags []string) string {
		return helper.NewProvider("clevercloud").
			SetOrganisation(tests.ORGANISATION).
			SetKeyValues(map[string]any{"default_tags": defaultTags}).
			Append(cellarBlock).
			String()
	}

	// resource-level tags never change across steps: provider tags must not leak in.
	resourceTags := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.StringExact("bar"),
		knownvalue.StringExact("foo"),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			// Merge: default_tags ∪ resource tags = tags_all.
			ResourceName: "cellar_" + rName,
			Config:       config([]string{"default"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), resourceTags),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags_all"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("default"),
					knownvalue.StringExact("foo"),
				})),
			},
		}, {
			// Propagation: adding a provider tag updates tags_all in place, tags unchanged.
			ResourceName: "cellar_" + rName,
			Config:       config([]string{"default", "platform"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), resourceTags),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags_all"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("default"),
					knownvalue.StringExact("foo"),
					knownvalue.StringExact("platform"),
				})),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}
