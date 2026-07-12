package materiakv_test

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

func TestAccMateriaKV_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-kv")
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_materia_kv.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).
		SetKeyValues(map[string]any{"default_tags": []string{"managed"}})
	materiakvBlock := helper.NewRessource("clevercloud_materia_kv", rName, helper.SetKeyValues(map[string]any{"name": rName, "region": "par", "tags": []string{"foo", "bar"}}))

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

// TestAccMateriaKV_tagsPropagation covers the default_tags lifecycle on an existing
// resource: a provider-level default_tags change must propagate to tags_all without
// touching the resource-level tags, and removing a resource-level tag must remove it
// from the API set.
func TestAccMateriaKV_tagsPropagation(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-kv-tags")
	fullName := fmt.Sprintf("clevercloud_materia_kv.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).
		SetKeyValues(map[string]any{"default_tags": []string{"managed"}})
	kvBlock := helper.NewRessource("clevercloud_materia_kv", rName, helper.SetKeyValues(map[string]any{"name": rName, "region": "par", "tags": []string{"foo", "bar"}}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// Create with default_tags ["managed"] and resource tags ["foo", "bar"].
			ResourceName: rName,
			Config:       providerBlock.Append(kvBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
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
			// Change ONLY the provider default_tags: the new default must propagate to
			// tags_all on the existing resource while the resource-level tags stay untouched.
			ResourceName: rName,
			Config: providerBlock.
				SetKeyValues(map[string]any{"default_tags": []string{"managed", "cost-center"}}).
				Append(kvBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("foo"),
				})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags_all"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("cost-center"),
					knownvalue.StringExact("foo"),
					knownvalue.StringExact("managed"),
				})),
			},
		}, {
			// Drop a resource-level tag: it must be removed from the API set while the
			// provider defaults are preserved.
			ResourceName: rName,
			Config:       providerBlock.Append(kvBlock.SetOneValue("tags", []string{"foo"})).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("foo"),
				})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags_all"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("cost-center"),
					knownvalue.StringExact("foo"),
					knownvalue.StringExact("managed"),
				})),
			},
		}},
	})
}
