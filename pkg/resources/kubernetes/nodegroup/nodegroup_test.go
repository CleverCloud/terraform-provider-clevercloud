package nodegroup_test

import (
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

func TestAccKubernetesNodegroup_basic(t *testing.T) {
	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes_nodegroup.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	k8sBlock := helper.
		NewRessource("clevercloud_kubernetes", rName).
		SetOneValue("name", rName)
	nodegroupBlock := helper.
		NewRessource(
			"clevercloud_kubernetes_nodegroup",
			rName,
			helper.SetKeyValues(map[string]any{
				"kubernetes_id": "${clevercloud_kubernetes." + rName + ".id}",
				"flavor":        "S",
				"name":          rName,
				// Single node keeps the test under the K8S product quota
				// when this and TestAccKubernetes_basic run in parallel:
				// 8G CP + 12G S-node = 20G here, +8G for the basic test = 28G
				// total against a 52G quota. Bumping size also bumps the wall
				// clock, so leave scaling coverage to a dedicated test.
				"size": 1,
			}),
		)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(k8sBlock, nodegroupBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				//statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^node_group_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("size"), knownvalue.Int64Exact(1)),
			},
		}},
	})
}
