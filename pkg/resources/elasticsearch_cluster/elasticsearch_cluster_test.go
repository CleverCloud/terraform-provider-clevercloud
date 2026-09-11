package elasticsearch_cluster_test

import (
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

func TestAccElasticsearchCluster_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-es")
	fullName := fmt.Sprintf("clevercloud_elasticsearch_cluster.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	esBlock := helper.NewRessource(
		"clevercloud_elasticsearch_cluster",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":       rName,
			"node_count": 3,
			"plan":       "M",
			"version": map[string]any{
				"major": 8,
			},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(esBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("node_count"), knownvalue.Int64Exact(3)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.StringExact("M")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version").AtMapKey("major"), knownvalue.Int64Exact(8)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version").AtMapKey("minor"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version").AtMapKey("patch"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("networkgroup_id"), knownvalue.StringRegexp(regexp.MustCompile(`^ng_`))),
				// the cluster is only reachable through its network group
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("endpoint"), knownvalue.StringRegexp(regexp.MustCompile(`\.cc-ng\.cloud$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("username"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("password"), knownvalue.NotNull()),
			},
		}},
	})
}

func TestAccElasticsearchCluster_versionValidation(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tf-test-es")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// minor without major should fail
	esBlock := helper.NewRessource(
		"clevercloud_elasticsearch_cluster",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":       rName,
			"node_count": 3,
			"plan":       "M",
			"version": map[string]any{
				"minor": 19,
			},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			Config:      providerBlock.Append(esBlock).String(),
			ExpectError: regexp.MustCompile(`Cannot set minor version without major version`),
		}},
	})
}

func TestAccElasticsearchCluster_unavailableVersion(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tf-test-es")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	esBlock := helper.NewRessource(
		"clevercloud_elasticsearch_cluster",
		rName,
		helper.SetKeyValues(map[string]any{
			"name": rName,
			"plan": "M",
			"version": map[string]any{
				"major": 99,
				"minor": 1,
				"patch": 1,
			},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			Config:      providerBlock.Append(esBlock).String(),
			ExpectError: regexp.MustCompile(`version 99\.1\.1 is not available, supported versions:`),
		}},
	})
}
