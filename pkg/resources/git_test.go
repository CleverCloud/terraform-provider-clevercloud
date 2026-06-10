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
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// This is a test for local Git repositories, we don't care about the runtime
func TestAccPython_localGit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-python")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

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
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(pythonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				// Test the state for provider's populated values
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
			},
		}},
	})
}

// Changing only environment variables on an application deployed from a local
// repository without a pinned commit must trigger a new deployment (restart):
// the no-op git push used to be counted as a deployment, so the restart was
// skipped and the application kept running with the old environment.
func TestAccApplication_envChangeTriggersNewDeployment(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-python")
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	repoDir := path.Join(os.TempDir(), "tfenvchangerepo")
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

	_, err = workTree.Commit("Initial commit", &git.CommitOptions{
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
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"environment":        map[string]any{"MY_KEY": "first"},
		}),
		// no commit: the repository HEAD is the deployment target
		helper.SetBlockValues("deployment", map[string]any{"repository": fmt.Sprintf("file://%s", repoDir)}),
	)

	// expectDeployments waits for at least count deployments to exist and for
	// none of them to still be running, so the next step can safely restart the app
	expectDeployments := func(count int) statecheck.StateCheck {
		return tests.NewCheckRemoteResource(
			fullName,
			func(ctx context.Context, id string) (*[]tmp.DeploymentResponse, error) {
				deadline := time.Now().Add(5 * time.Minute)
				for {
					deploymentsRes := tmp.ListDeployments(ctx, cc, tests.ORGANISATION, id)
					if deploymentsRes.HasError() {
						return nil, deploymentsRes.Error()
					}
					deployments := *deploymentsRes.Payload()

					running := pkg.HasSome(deployments, func(d tmp.DeploymentResponse) bool {
						return d.State == "WIP"
					})
					if len(deployments) >= count && !running {
						return &deployments, nil
					}

					if time.Now().After(deadline) {
						return nil, fmt.Errorf("expected %d finished deployments, got %d (running: %t)", count, len(deployments), running)
					}
					time.Sleep(10 * time.Second)
				}
			},
			func(ctx context.Context, id string, state *tfjson.State, deployments *[]tmp.DeploymentResponse) error {
				return nil
			})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// initial apply: the git push triggers the first deployment
			ResourceName: rName,
			Config:       providerBlock.Append(pythonBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				expectDeployments(1),
			},
		}, {
			// same code, only an env var changes: a restart must trigger a second deployment
			ResourceName: rName,
			Config: providerBlock.Append(
				pythonBlock.SetOneValue("environment", map[string]any{"MY_KEY": "second"}),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				expectDeployments(2),
				tests.NewCheckRemoteResource(
					fullName,
					func(ctx context.Context, id string) (*[]tmp.Env, error) {
						envRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
						if envRes.HasError() {
							return nil, envRes.Error()
						}
						return envRes.Payload(), nil
					},
					func(ctx context.Context, id string, state *tfjson.State, envs *[]tmp.Env) error {
						env := pkg.First(*envs, func(e tmp.Env) bool { return e.Name == "MY_KEY" })
						if env == nil || env.Value != "second" {
							return tests.AssertError("bad env var value MY_KEY", env, "second")
						}
						return nil
					}),
			},
		}},
	})
}
