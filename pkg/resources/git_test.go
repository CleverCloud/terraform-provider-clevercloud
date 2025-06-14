package resources_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"regexp"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

// This is a test for local Git repositories, we don't care about the runtime
func TestAccPython_localGit(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-python-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)

	repoDir := path.Join(os.TempDir(), "tfsamplerepo")
	os.RemoveAll(repoDir)       // clean old instance before test
	defer os.RemoveAll(repoDir) // clean after test

	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatalf("failed to initialize test repository: %s", err)
	}

	err = os.WriteFile(path.Join(repoDir, "README.md"), []byte("# Test repository"), 0644)
	if err != nil {
		t.Fatalf("failed to write README.md: %s", err)
	}
	workTree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %s", err)
	}

	_, err = workTree.Add("README.md")
	if err != nil {
		t.Fatalf("failed to add README.md: %s", err)
	}

	hash, err := workTree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Terraform",
			Email: "terraform@localhost",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit README.md: %s", err)
	}

	pythonBlock := helper.NewRessource(
		"clevercloud_python",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"redirect_https":     true,
			"sticky_sessions":    true,
		}),
		helper.SetBlockValues(
			"deployment",
			map[string]any{"repository": fmt.Sprintf("file://%s", repoDir), "commit": hash.String()},
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetApp(ctx, cc, org, resource.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}
				if res.Payload().State == "TO_DELETE" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted state: '%s'", resource.Primary.ID, res.Payload().State)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(pythonBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				// Test the state for provider's populated values
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^app_.*$`)),
				// Test CleverCloud API for configured applications
				func(state *terraform.State) error {
					//id := state.RootModule().Resources[fullName].Primary.ID

					/*appRes := tmp.GetApp(ctx, cc, org, id)
					if appRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appRes.Error())
					}
					app := appRes.Payload()*/
					return nil
				},
			),
		}},
	})
}
