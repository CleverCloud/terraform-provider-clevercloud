package application_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// TestAccTags_application checks the tags lifecycle on an application:
// set on create, replaced on update, left untouched when removed from the
// configuration, and cleared with an empty set.
func TestAccTags_application(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-tags")
	fullName := fmt.Sprintf("clevercloud_nodejs.%s", rName)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	appBlock := helper.NewRessource(
		"clevercloud_nodejs",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"tags":               []string{"terraform", "env:test"},
		}),
	)

	fetchTags := func(ctx context.Context, id string) (*[]string, error) {
		res := tmp.GetAppTags(ctx, cc, tests.ORGANISATION, id)
		if res.HasError() {
			return nil, res.Error()
		}
		return res.Payload(), nil
	}
	expectRemoteTags := func(expected ...string) statecheck.StateCheck {
		return tests.NewCheckRemoteResource(fullName, fetchTags, func(ctx context.Context, id string, state *tfjson.State, tags *[]string) error {
			got := slices.Clone(*tags)
			slices.Sort(got)
			slices.Sort(expected)
			if !slices.Equal(got, expected) {
				return fmt.Errorf("remote tags mismatch: got %v, want %v", got, expected)
			}
			return nil
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			Config: providerBlock.Append(appBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("terraform"),
					knownvalue.StringExact("env:test"),
				})),
				expectRemoteTags("terraform", "env:test"),
			},
		}, {
			Config: providerBlock.Append(appBlock.SetOneValue("tags", []string{"terraform"})).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("terraform"),
				})),
				expectRemoteTags("terraform"),
			},
		}, {
			// Removing the attribute stops managing tags, the remote ones stay
			Config: providerBlock.Append(appBlock.UnsetOneValue("tags")).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.Null()),
				expectRemoteTags("terraform"),
			},
		}, {
			// An empty set clears the remote tags
			Config: providerBlock.Append(appBlock.SetOneValue("tags", []string{})).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{})),
				expectRemoteTags(),
			},
		}},
	})
}
