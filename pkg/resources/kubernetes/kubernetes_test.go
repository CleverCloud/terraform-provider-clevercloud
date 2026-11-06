package kubernetes_test

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
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
	// terraform_data is only created during the apply, so its output is unknown
	// at plan time and so is the attribute that reads it. Reading the plan would
	// silently mean false: no Karpenter on a create, and a request to uninstall
	// it on an update, both followed by an apply failing on the unknown written
	// back to state
	flagBlock := func(enabled bool) *helper.Ressource {
		return helper.NewRessource("terraform_data", "flag", helper.SetKeyValues(map[string]any{"input": enabled}))
	}
	unknownBlock := helper.NewRessource(
		"clevercloud_kubernetes",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":                  rName,
			"node_autoprovisioning": helper.Expression("terraform_data.flag.output"),
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// Karpenter is installed on creation, from a value the plan does not know
			ResourceName: rName,
			Config:       providerBlock.Append(flagBlock(true), unknownBlock).String(),
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
		}, {
			// an update whose planned value is unknown must leave the feature
			// alone rather than read the unknown as a request to disable it
			ResourceName: rName,
			Config:       providerBlock.Append(flagBlock(true), unknownBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("node_autoprovisioning"), knownvalue.Bool(true)),
			},
		}},
	})
}

func TestAccKubernetes_options(t *testing.T) {
	// the versions are fetched before resource.Test gets a chance to skip
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC is not set")
	}

	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// the platform decides which versions exist, the test must outlive them
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))
	product := sdk.NewSDK(sdk.WithClient(cc)).V4().KubernetesProduct().Getkubernetesserviceconfig(ctx)
	if product.HasError() {
		t.Fatalf("failed to get the kubernetes service configuration: %s", product.Error())
	}
	available := product.Payload().Versions.Available
	if len(available) < 2 {
		t.Skipf("an upgrade takes two versions, the platform offers %v", available)
	}
	newest, older := available[0], available[1]

	cluster := func(values map[string]any) string {
		return providerBlock.Append(helper.NewRessource("clevercloud_kubernetes", rName, helper.SetKeyValues(values))).String()
	}
	allInOne := map[string]any{"type": "ALL_IN_ONE", "flavor": "S", "replication_factor": 1}
	component := map[string]any{"flavor": "2XS", "replication_factor": 1}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// refused values fail the plan, nothing gets created
			ResourceName: rName,
			Config:       cluster(map[string]any{"name": rName, "version": "0.1"}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`version '0.1' is not available`),
		}, {
			ResourceName: rName,
			Config:       cluster(map[string]any{"name": rName, "topology": map[string]any{"type": "ALL_IN_ONE", "flavor": "2XS", "replication_factor": 1}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`not allowed for ALL_IN_ONE`),
		}, {
			ResourceName: rName,
			Config:       cluster(map[string]any{"name": rName, "topology": map[string]any{"type": "ALL_IN_ONE", "flavor": "s", "replication_factor": 1}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`uppercase`),
		}, {
			ResourceName: rName,
			Config:       cluster(map[string]any{"name": rName, "topology": map[string]any{"type": "DISTRIBUTED"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`missing topology components`),
		}, {
			ResourceName: rName,
			Config: cluster(map[string]any{
				"name":               rName,
				"version":            older,
				"persistent_storage": true,
				"tags":               []string{"tf-test", "env:test"},
				"topology":           allInOne,
			}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.StringExact(older)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("persistent_storage"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("tf-test"), knownvalue.StringExact("env:test")})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("type"), knownvalue.StringExact("ALL_IN_ONE")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("flavor"), knownvalue.StringExact("S")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("replication_factor"), knownvalue.Int64Exact(1)),
			},
		}, {
			// everything but the topology changes in place
			ResourceName: rName,
			Config: cluster(map[string]any{
				"name":               rName + "-renamed",
				"version":            newest,
				"persistent_storage": false,
				"tags":               []string{},
				"topology":           allInOne,
			}),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(fullName, plancheck.ResourceActionUpdate)},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName+"-renamed")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.StringExact(newest)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("persistent_storage"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetSizeExact(0)),
			},
		}, {
			// an older version is the same in place operation
			ResourceName: rName,
			Config:       cluster(map[string]any{"name": rName + "-renamed", "version": older, "tags": []string{}, "topology": allInOne}),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(fullName, plancheck.ResourceActionUpdate)},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.StringExact(older)),
			},
		}, {
			// removing the attributes keeps the values and plans nothing
			ResourceName: rName,
			Config:       cluster(map[string]any{"name": rName + "-renamed", "tags": []string{}}),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.StringExact(older)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("type"), knownvalue.StringExact("ALL_IN_ONE")),
			},
		}, {
			ResourceName: rName,
			Config: cluster(map[string]any{
				"name": rName + "-renamed",
				"topology": map[string]any{
					"type": "DISTRIBUTED",
					"components": map[string]any{
						"apiserver":                component,
						"cloud_controller_manager": component,
						"controller_manager":       component,
						"node_group_operator":      component,
						"scheduler":                component,
					},
				},
			}),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(fullName, plancheck.ResourceActionReplace)},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("type"), knownvalue.StringExact("DISTRIBUTED")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("flavor"), knownvalue.Null()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("components").AtMapKey("scheduler").AtMapKey("flavor"), knownvalue.StringExact("2XS")),
			},
		}, {
			ResourceName:            fullName,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"kubeconfig", "tags"},
		}},
	})
}

// TestAccKubernetes_upgradeState creates a cluster with a release whose state
// predates version and topology, then plans its real topology with the current
// code, without any refresh. Only UpgradeState can fill the state at that point,
// and planning nothing proves the cluster is not replaced
func TestAccKubernetes_upgradeState(t *testing.T) {
	// the state records the provider address, it has to be the released one
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "CleverCloud")

	ctx := t.Context()
	rName := fmt.Sprintf("tf-test-kubernetes-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_kubernetes.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	cluster := func(values map[string]any) string {
		values["name"] = rName
		return providerBlock.Append(helper.NewRessource("clevercloud_kubernetes", rName, helper.SetKeyValues(values))).String()
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     tests.ExpectOrganisation(t),
		CheckDestroy: tests.CheckDestroy(ctx),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{{
			ExternalProviders: map[string]resource.ExternalProvider{
				"clevercloud": {Source: "CleverCloud/clevercloud", VersionConstraint: "3.0.0"},
			},
			Config: cluster(map[string]any{}),
		}, {
			ProtoV6ProviderFactories: tests.ProtoV6Provider,
			Config: cluster(map[string]any{
				"topology": map[string]any{"type": "DEDICATED_COMPUTE", "flavor": "XS", "replication_factor": 1},
			}),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(fullName, plancheck.ResourceActionNoop)},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("persistent_storage"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("type"), knownvalue.StringExact("DEDICATED_COMPUTE")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("topology").AtMapKey("flavor"), knownvalue.StringExact("XS")),
			},
		}},
	})
}
