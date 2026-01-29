package networkgroup_test

import (
	"context"
	_ "embed"
	"fmt"
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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccNG_basic(t *testing.T) {
	t.Parallel()
	rName := acctest.RandomWithPrefix("tf-test-ng")
	fullName := fmt.Sprintf("clevercloud_networkgroup.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	addonBlock := helper.NewRessource(
		"clevercloud_networkgroup",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":        rName,
			"description": "par",
			"tags":        []string{"tag1", "tag2"},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetNetworkgroup(context.Background(), cc, tests.ORGANISATION, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect resource '%s' to be deleted", resource.Primary.ID)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^ng_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("description"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("tag1"),
					knownvalue.StringExact("tag2"),
				})),
			},
		}},
	})
}

// TestAccNG_withPeers issue #337
// When a network group has peers (running instances), terraform plan/refresh
// fails with: "cannot parse response body: json: cannot unmarshal object into Go struct field NetworkGroup1.peers of type models.Peer"
func TestAccNG_withPeers(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ngName := acctest.RandomWithPrefix("tf-test-ng")
	appName := acctest.RandomWithPrefix("tf-test-app")
	ngFullName := fmt.Sprintf("clevercloud_networkgroup.%s", ngName)
	appFullName := fmt.Sprintf("clevercloud_docker.%s", appName)
	cc := client.New(client.WithAutoOauthConfig())

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	ngBlock := helper.NewRessource("clevercloud_networkgroup", ngName, helper.SetKeyValues(map[string]any{
		"name":        ngName,
		"description": "Network group with peers test",
	}))

	appBlock := helper.NewRessource("clevercloud_docker", appName,
		helper.SetKeyValues(map[string]any{
			"name":               appName,
			"region":             "par",
			"min_instance_count": 1, // This will create instances (peers)
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"networkgroups": []map[string]any{{
				"networkgroup_id": fmt.Sprintf("${clevercloud_networkgroup.%s.id}", ngName),
				"fqdn":            fmt.Sprintf("%s.ng", appName),
			}},
		}),
		helper.SetBlockValues("deployment", map[string]any{
			"repository": "https://github.com/CleverCloud/rust-docker-example.git",
			"commit":     "24f206d760a35ae1847a61f35e496de0edbfb90b",
		}),
	)
	config := providerBlock.Append(ngBlock, appBlock).String()

	t.Logf("CONFIG:\n%s", config)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				if resource.Type == "clevercloud_networkgroup" {
					res := tmp.GetNetworkgroup(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
					if res.IsNotFoundError() {
						continue
					}
					if res.HasError() {
						return fmt.Errorf("unexpected error: %s", res.Error().Error())
					}
					return fmt.Errorf("expect networkgroup '%s' to be deleted", resource.Primary.ID)
				}
				if resource.Type == "clevercloud_docker" {
					res := tmp.GetApp(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
					if res.IsNotFoundError() {
						continue
					}
					if res.HasError() {
						return fmt.Errorf("unexpected error: %s", res.Error().Error())
					}
					if res.Payload().State == "TO_DELETE" {
						continue
					}
					return fmt.Errorf("expect application '%s' to be deleted", resource.Primary.ID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{{
			Config: config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(ngFullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^ng_.*`))),
				statecheck.ExpectKnownValue(ngFullName, tfjsonpath.New("name"), knownvalue.StringExact(ngName)),
				statecheck.ExpectKnownValue(appFullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*`))),
			},
			Check: func(state *terraform.State) error {
				// Get the network group ID and app ID from state
				var ngID, appID string
				for _, rs := range state.RootModule().Resources {
					if rs.Type == "clevercloud_networkgroup" {
						ngID = rs.Primary.ID
					}
					if rs.Type == "clevercloud_docker" {
						appID = rs.Primary.ID
					}
				}

				t.Logf("Waiting for application deployment to complete and instances to start...")

				// Wait for deployment to complete and instances to be running (up to 5 minutes)
				maxWait := 5 * time.Minute
				checkInterval := 10 * time.Second
				deadline := time.Now().Add(maxWait)

				var hasRunningInstances bool
				for time.Now().Before(deadline) {
					// Check if the application has running instances
					instancesRes := tmp.ListInstances(ctx, cc, tests.ORGANISATION, appID)
					if instancesRes.HasError() {
						t.Logf("Warning: Could not list instances: %v", instancesRes.Error())
					} else {
						instances := instancesRes.Payload()
						t.Logf("Application has %d instances", len(*instances))

						// Check if any instance is UP
						for _, inst := range *instances {
							if inst.State == "UP" {
								hasRunningInstances = true
								t.Logf("Found running instance: %s (state: %s)", inst.ID, inst.State)
								break
							}
						}

						if hasRunningInstances {
							break
						}
					}

					// Wait before checking again
					t.Logf("No running instances yet, waiting %v before retry...", checkInterval)
					time.Sleep(checkInterval)
				}

				if !hasRunningInstances {
					t.Logf("Warning: No running instances after %v.", maxWait)
				}

				// Try to get the network group using tmp client which should work
				t.Logf("=== ATTEMPTING TO READ NETWORK GROUP ===")
				ngRes := tmp.GetNetworkgroup(ctx, cc, tests.ORGANISATION, ngID)
				if ngRes.HasError() {
					t.Logf("✗ Failed to get network group: %v", ngRes.Error())
					t.Logf("This is the bug! The API response cannot be deserialized.")
				} else {
					ng := ngRes.Payload()
					t.Logf("✓ Successfully read network group via tmp client")
					t.Logf("  Members: %d", len(ng.Members))
					t.Logf("  Peers: %d", len(ng.Peers))

					// If we have peers, dump their structure
					if len(ng.Peers) > 0 {
						t.Logf("=== PEER DETAILS ===")
						for i, peer := range ng.Peers {
							t.Logf("Peer %d:", i)
							t.Logf("  ID: %s", peer.ID)
							t.Logf("  Label: %s", peer.Label)
							t.Logf("  PublicKey: %s", peer.PublicKey)
							t.Logf("  Hostname: %s", peer.Hostname)
							t.Logf("  ParentMember: %s", peer.ParentMember)
							t.Logf("  ParentEvent: %s", peer.ParentEvent)
							t.Logf("  HV: %s", peer.HV)
							t.Logf("  Endpoint.IP: %s", peer.Endpoint.IP)
						}
					}
				}
				t.Logf("=== END NETWORK GROUP READ ===")

				return nil
			},
		}, {
			// This RefreshState step should trigger the bug if peers exist
			// The provider's Read function uses the SDK which has incorrect peer types
			RefreshState: true,
		}},
	})
}
