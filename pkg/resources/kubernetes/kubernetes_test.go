package kubernetes_test

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

func TestAccKubernetes_basic(t *testing.T) {
	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	kubernetesBlock := helper.NewRessource(
		"clevercloud_kubernetes",
		rName,
		helper.SetKeyValues(map[string]any{
			"name": rName,
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(kubernetesBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectIdentityValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^kubernetes_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("node_autoprovisioning"), knownvalue.Bool(false)),
			},
		}},
	})
}

func TestAccKubernetes_nodeAutoprovisioning(t *testing.T) {
	t.Skip("the node auto-provisioning route is not deployed yet, see axo MR !1556")

	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	kubernetesBlock := func(enabled bool) *helper.Ressource {
		return helper.NewRessource(
			"clevercloud_kubernetes",
			rName,
			helper.SetKeyValues(map[string]any{
				"name":                  rName,
				"node_autoprovisioning": enabled,
			}),
		)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// Karpenter is installed on creation
			ResourceName: rName,
			Config:       providerBlock.Append(kubernetesBlock(true)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectIdentityValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^kubernetes_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("node_autoprovisioning"), knownvalue.Bool(true)),
			},
		}, {
			// then uninstalled in place, without replacing the cluster
			ResourceName: rName,
			Config:       providerBlock.Append(kubernetesBlock(false)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("node_autoprovisioning"), knownvalue.Bool(false)),
			},
		}, {
			// and installed again on an existing cluster
			ResourceName: rName,
			Config:       providerBlock.Append(kubernetesBlock(true)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("node_autoprovisioning"), knownvalue.Bool(true)),
			},
		}},
	})
}
