package elasticsearch_cluster_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
			"name":        rName,
			"node_count":  3,
			"cpu_count":   1,
			"memory_size": 1024,
			"disk_size":   1024,
			"version": map[string]any{
				"major": 8,
				"minor": 19,
				"patch": 9,
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
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("cpu_count"), knownvalue.Int64Exact(1)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("memory_size"), knownvalue.Int64Exact(1024)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("disk_size"), knownvalue.Int64Exact(1024)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("version").AtMapKey("major"), knownvalue.Int64Exact(8)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("endpoint"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("username"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("password"), knownvalue.NotNull()),
			},
			Check: waitForElasticsearchHealthy(ctx, fullName, 10*time.Minute),
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
			"name":        rName,
			"node_count":  1,
			"cpu_count":   1,
			"memory_size": 1024,
			"disk_size":   1024,
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

type esClusterHealth struct {
	ClusterName   string `json:"cluster_name"`
	Status        string `json:"status"`
	NumberOfNodes int    `json:"number_of_nodes"`
}

// waitForElasticsearchHealthy polls the cluster's /_cluster/health endpoint
// until it reports green or yellow status, confirming the cluster is operational.
func waitForElasticsearchHealthy(ctx context.Context, resourceFullName string, timeout time.Duration) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceFullName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceFullName)
		}

		endpoint := rs.Primary.Attributes["endpoint"]
		username := rs.Primary.Attributes["username"]
		password := rs.Primary.Attributes["password"]

		if endpoint == "" || username == "" || password == "" {
			return fmt.Errorf("missing endpoint/username/password in state for %q", resourceFullName)
		}

		healthURL := fmt.Sprintf("https://%s/_cluster/health", endpoint)
		httpClient := &http.Client{Timeout: 10 * time.Second}

		deadline, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		for {
			select {
			case <-deadline.Done():
				return fmt.Errorf("timeout waiting for Elasticsearch cluster %q to be healthy", resourceFullName)
			default:
				httpReq, err := http.NewRequestWithContext(deadline, http.MethodGet, healthURL, nil)
				if err != nil {
					return fmt.Errorf("failed to create request: %w", err)
				}
				httpReq.SetBasicAuth(username, password)

				httpResp, err := httpClient.Do(httpReq)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}

				var health esClusterHealth
				decodeErr := json.NewDecoder(httpResp.Body).Decode(&health)
				httpResp.Body.Close()

				if decodeErr != nil || httpResp.StatusCode != http.StatusOK {
					time.Sleep(5 * time.Second)
					continue
				}

				if health.Status == "green" || health.Status == "yellow" {
					fmt.Printf("[INFO] Elasticsearch cluster is healthy: status=%s, nodes=%d, name=%s\n",
						health.Status, health.NumberOfNodes, health.ClusterName)
					return nil
				}

				time.Sleep(5 * time.Second)
			}
		}
	}
}
