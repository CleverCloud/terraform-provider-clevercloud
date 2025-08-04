package nodegroup_test

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccKubernetesNodegroup_basic(t *testing.T) {
	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes_nodegroup.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
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
		CheckDestroy: func(state *terraform.State) error {
			for resourceName, resource := range state.RootModule().Resources {
				if strings.HasPrefix(resourceName, "clevercloud_kubernetes_nodegroup") {
					// Get the kubernetes_id from the nodegroup resource attributes
					kubernetesID := resource.Primary.Attributes["kubernetes_id"]
					if kubernetesID == "" {
						continue
					}

					// Check if the kubernetes cluster still exists
					// If it doesn't, the nodegroup is also gone
					k8sRes := tmp.GetKubernetes(ctx, cc, tests.ORGANISATION, kubernetesID)
					if k8sRes.IsNotFoundError() {
						continue
					}

					// TODO: issue #704
					for range 30 {
						res := tmp.ListNodeGroups(ctx, cc, tests.ORGANISATION, kubernetesID)
						if res.IsNotFoundError() {
							continue
						}
						if res.HasError() {
							return fmt.Errorf("unexpected error getting node group: %s", res.Error().Error())
						}
						groups := *res.Payload()

						if len(groups) == 0 || groups[0].Status == "DELETED" {
							return nil
						}
						t.Logf("NodeGroups: %+v", groups)
						time.Sleep(10 * time.Second)
					}
					return fmt.Errorf("expect nodegroup to be deleted")
				}
			}
			return nil
		},
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
