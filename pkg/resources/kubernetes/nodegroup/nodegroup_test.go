package nodegroup_test

import (
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/kubernetes"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccKubernetes_basic(t *testing.T) {
	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes_nodegroup.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	k8sRes := tmp.CreateKubernetes(ctx, cc, tests.ORGANISATION, tmp.KubernetesCreateRequest{
		Name: "tf-test-kuberntest-nodegroupdep",
	})
	if k8sRes.HasError() {
		t.Fatalf("failed to create dependencies: %s", k8sRes.Error().Error())
	}
	k8s := k8sRes.Payload()

	t.Cleanup(func() {
		deleteRes := tmp.DeleteKubernetes(ctx, cc, tests.ORGANISATION, k8s.ID)
		if deleteRes.HasError() {
			t.Errorf("failed to delete cluster: %s", deleteRes.Error().Error())
		}
	})

	for state := range kubernetes.WaitForKubernetes(ctx, cc, tests.ORGANISATION, k8s.ID, 1*time.Second) {
		t.Logf("Cluster state changed: %s", state.Status)
	}

	kubernetesBlock := helper.NewRessource(
		"clevercloud_kubernetes_nodegroup",
		rName,
		helper.SetKeyValues(map[string]any{
			"kubernetes_id": k8s.ID,
			"flavor":        "xs",
			"name":          rName,
			"size":          3,
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				fmt.Printf("\n\n%+v\n\n", resource.Primary)
				res := tmp.ListNodeGroups(ctx, cc, tests.ORGANISATION, k8s.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error getting node group: %s", res.Error().Error())
				}

				for _, ng := range *res.Payload() {
					if ng.Name == rName {
						return fmt.Errorf("expect resource '%s' to be deleted", ng.Name)
					}
				}

			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			//PreConfig: func() {	},
			Config: providerBlock.Append(kubernetesBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				//statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^node_group_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("size"), knownvalue.Int64Exact(3)),
			},
		}},
	})
}
