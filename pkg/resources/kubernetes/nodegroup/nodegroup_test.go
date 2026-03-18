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

	nodegroupBlock := helper.NewRessource(
		"clevercloud_kubernetes_nodegroup",
		rName,
		helper.SetKeyValues(map[string]any{
			"kubernetes_id": "${clevercloud_kubernetes." + rName + ".id}",
			"flavor":        "XS",
			"name":          rName,
			"size":          3,
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			//PreConfig: func() {	},
			Config: providerBlock.Append(k8sBlock, nodegroupBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				//statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^node_group_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("size"), knownvalue.Int64Exact(3)),
			},
		}},
	})
}
