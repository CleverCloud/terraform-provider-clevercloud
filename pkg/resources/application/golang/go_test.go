package golang_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/app"
	"go.clever-cloud.dev/sdk/models"
)

func TestAccGo_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-go")
	fullName := fmt.Sprintf("clevercloud_go.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	goBlock := helper.NewRessource(
		"clevercloud_go",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"build_flavor":       "M",
			"redirect_https":     true,
			"sticky_sessions":    true,
			"app_folder":         "./app",
			"environment":        map[string]any{"MY_KEY": "myval"},
			"dependencies":       []string{},
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
	)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(goBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("M")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("vhosts"), knownvalue.SetExact(
					[]knownvalue.Check{
						knownvalue.ObjectExact(
							map[string]knownvalue.Check{
								"fqdn":       knownvalue.StringRegexp(regexp.MustCompile(`^app-.*\.cleverapps\.io$`)),
								"path_begin": knownvalue.StringExact("/"),
							},
						),
					},
				)),
				tests.NewCheckRemoteResource(
					fullName,
					func(ctx context.Context, id string) (*models.ApplicationView, error) {
						appRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Getorganisationapplication(ctx)
						if appRes.HasError() {
							return nil, appRes.Error()
						}
						return appRes.Payload(), nil
					},
					func(ctx context.Context, id string, state *tfjson.State, appRes *models.ApplicationView) error {
						if appRes.Name != rName {
							return tests.AssertError("invalid name", appRes.Name, rName)
						}

						if appRes.Instance.MinInstances != 1 {
							return tests.AssertError("invalid min instance count", appRes.Instance.MinInstances, "1")
						}

						if appRes.Instance.MaxInstances != 2 {
							return tests.AssertError("invalid name", appRes.Name, rName)
						}

						if pkg.FlavorName(appRes.Instance.MinFlavor) != "XS" {
							return tests.AssertError("invalid name", appRes.Name, rName)
						}

						if pkg.FlavorName(appRes.Instance.MaxFlavor) != "M" {
							return tests.AssertError("invalid name", appRes.Name, rName)
						}

						if appRes.ForceHTTPS != "ENABLED" {
							return tests.AssertError("expect option to be set", "redirect_https", appRes.ForceHTTPS)
						}

						if !appRes.StickySessions {
							return tests.AssertError("expect option to be set", "sticky_sessions", appRes.StickySessions)
						}
						if appRes.Zone != "par" {
							return tests.AssertError("expect region to be 'par'", "region", appRes.Zone)
						}

						if len(appRes.Vhosts) != 1 {
							return tests.AssertError("expect one vhost", appRes.Vhosts, "<cleverapps>")
						}

						if !strings.HasSuffix(appRes.Vhosts[0].Fqdn, ".cleverapps.io/") {
							return tests.AssertError("expect a cleverapps fqdn", appRes.Vhosts[0].Fqdn, "<cleverapps>")
						}

						if len(appRes.Vhosts) != 1 {
							return tests.AssertError("expect one vhost", appRes.Vhosts, "<cleverapps>")
						}

						if !strings.HasSuffix(appRes.Vhosts[0].Fqdn, ".cleverapps.io/") {
							return tests.AssertError("expect a cleverapps fqdn", appRes.Vhosts[0].Fqdn, "<cleverapps>")
						}

						appEnvRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Env().Listorganisationapplicationenvironmentvariables(ctx)
						if appEnvRes.HasError() {
							return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
						}

						env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e models.AddonEnvironmentView) map[string]string {
							acc[e.Name] = e.Value
							return acc
						})

						v := env["MY_KEY"]
						if v != "myval" {
							return tests.AssertError("bad env var value MY_KEY", "myval3", v)
						}

						v2 := env[app.APP_FOLDER]
						if v2 != "./app" {
							return tests.AssertError("bad env var value APP_FOLER", "./app", v2)
						}

						v3 := env[app.CC_POST_BUILD_HOOK]
						if v3 != "echo \"build is OK!\"" {
							return tests.AssertError("bad env var value CC_POST_BUILD_HOOK", "echo \"build is OK!\"", v3)
						}
						return nil
					}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				goBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}},
	})
}

// waitForDeploymentComplete waits for the most recent deployment to reach a terminal state (OK or FAIL)
func waitForDeploymentComplete(t *testing.T, cc *client.Client, orgID, appID string) error {
	ctx := t.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for deployment to complete")
		case <-ticker.C:
			deploymentsRes := pkg.SDK(cc).V2().Organisations().ID(orgID).Applications().Appid(appID).Deployments().Getapplicationdeploymentsfororga(ctx)
			if deploymentsRes.IsNotFoundError() {
				continue
			}
			if deploymentsRes.HasError() {
				return fmt.Errorf("failed to list deployments: %w", deploymentsRes.Error())
			}

			deployments := *deploymentsRes.Payload()
			if len(deployments) == 0 {
				continue
			}

			// Check the most recent deployment (first in the list)
			deploying := false
			for _, deployment := range deployments {
				t.Logf("DEPLOY\t%s\t%s", deployment.UUID, deployment.State)
				if deployment.State != "OK" && deployment.State != "FAIL" {
					deploying = true
				}
			}
			if !deploying {
				return nil
			}
		}
	}
}

// goServer returns a main.go serving version on port 8080.
func goServer(version string) string {
	return `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "` + version + `")
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
`
}

// TestAccGo_singleDeploymentOnEnvAndGitChange tests that when both env vars and git commit change,
// only a single deployment is triggered (not two separate deployments)
// This is a regression test for https://github.com/CleverCloud/terraform-provider-clevercloud/issues/179
func TestAccGo_singleDeploymentOnEnvAndGitChange(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-go-deploy")
	fullName := fmt.Sprintf("clevercloud_go.%s", rName)

	// Create a temporary directory for the git repo
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize git repo with first commit
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create initial files and commit: a module listening on 8080, so both
	// deployments succeed instead of waiting for the health check to give up
	if err := os.WriteFile(filepath.Join(repoPath, "go.mod"), []byte("module example.com/app\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	if _, err := worktree.Add("go.mod"); err != nil {
		t.Fatalf("failed to add go.mod: %v", err)
	}

	mainFile := filepath.Join(repoPath, "main.go")
	if err := os.WriteFile(mainFile, []byte(goServer("v1")), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	if _, err := worktree.Add("main.go"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	commit1, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	t.Logf("Initial commit: %s", commit1)

	// Create second commit upfront (before test runs)
	if err := os.WriteFile(mainFile, []byte(goServer("v2")), 0644); err != nil {
		t.Fatalf("failed to update main.go: %v", err)
	}

	if _, err := worktree.Add("main.go"); err != nil {
		t.Fatalf("failed to add updated file: %v", err)
	}

	commit2, err := worktree.Commit("Update to v2", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit v2: %v", err)
	}
	t.Logf("Secondary commit: %s", commit2)

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	var initialDeploymentCount int
	var appID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// Step 1: Create an app with initial commit and env var
			ResourceName: rName,
			Config: providerBlock.Append(
				helper.NewRessource(
					"clevercloud_go",
					rName,
					helper.SetKeyValues(map[string]any{
						"name":               rName,
						"region":             "par",
						"min_instance_count": 1,
						"max_instance_count": 1,
						"smallest_flavor":    "XS",
						"biggest_flavor":     "XS",
						"environment":        map[string]any{"VERSION": "v1"},
					}),
					helper.SetBlockValues("deployment", map[string]any{
						"repository": "file://" + repoPath,
						"commit":     commit1.String(),
					}),
				),
			).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				func(state *terraform.State) error {
					rs, ok := state.RootModule().Resources[fullName]
					if !ok {
						return fmt.Errorf("resource not found: %s", fullName)
					}

					appID = rs.Primary.ID

					if err := waitForDeploymentComplete(t, cc, tests.ORGANISATION, appID); err != nil {
						t.Fatalf("failed waitin for deployment to end: %s", err.Error())
					}

					// Get initial deployment count
					deploymentsRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(appID).Deployments().Getapplicationdeploymentsfororga(ctx)
					if deploymentsRes.HasError() {
						return fmt.Errorf("failed to list deployments: %w", deploymentsRes.Error())
					}
					deployments := *deploymentsRes.Payload()

					initialDeploymentCount = len(deployments)
					t.Logf("Initial deployment count: %d", initialDeploymentCount)

					if initialDeploymentCount != 1 {
						return fmt.Errorf("expected at least one deployment, got %d", initialDeploymentCount)
					}

					return nil
				},
			),
		}, {
			// Step 2: Change BOTH env var AND git commit in the same terraform apply
			ResourceName: rName,
			Config: providerBlock.Append(
				helper.NewRessource(
					"clevercloud_go",
					rName,
					helper.SetKeyValues(map[string]any{
						"name":               rName,
						"region":             "par",
						"min_instance_count": 1,
						"max_instance_count": 1,
						"smallest_flavor":    "XS",
						"biggest_flavor":     "XS",
						"environment":        map[string]any{"VERSION": "v2"},
					}),
					helper.SetBlockValues("deployment", map[string]any{
						"repository": "file://" + repoPath,
						"commit":     commit2.String(),
					}),
				),
			).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				func(state *terraform.State) error {
					rs, ok := state.RootModule().Resources[fullName]
					if !ok {
						return fmt.Errorf("resource not found: %s", fullName)
					}

					appID := rs.Primary.ID

					time.Sleep(2 * time.Minute) // wait for the next deployment to appear (and unwanted too)
					if err := waitForDeploymentComplete(t, cc, tests.ORGANISATION, appID); err != nil {
						return fmt.Errorf("failed to wait for deployments to end: %s", err.Error())
					}
					// Get new deployment count
					deploymentsRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(appID).Deployments().Getapplicationdeploymentsfororga(ctx)
					if deploymentsRes.HasError() {
						return fmt.Errorf("failed to list deployments: %w", deploymentsRes.Error())
					}
					deployments := *deploymentsRes.Payload()

					newDeploymentCount := len(deployments)
					t.Logf("New deployment count: %d", newDeploymentCount)

					if newDeploymentCount != 2 {
						return fmt.Errorf(
							"expected exactly %d Git deployment(s), but got %d new. This suggests env change and git change triggered separate deployments (bug #179)",
							2, newDeploymentCount)
					}

					for _, deployment := range deployments {
						if deployment.State != "OK" {
							return fmt.Errorf("deployment %s ended in %s, want OK", deployment.UUID, deployment.State)
						}
					}

					return nil
				},
			),
		},
		},
	})
}

func TestAccGo_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-go-env")
	fullName := fmt.Sprintf("clevercloud_go.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// config builds each step's configuration from scratch: SetOneValue
	// mutates the block, and steps must not leak values into each other.
	config := func(overrides map[string]any) string {
		values := map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"build_tool":         "makefile",
			"binary":             "bin/server",
			"package":            "example.com/app/cmd/server",
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_go", rName, helper.SetKeyValues(values))).String()
	}

	remoteEnv := func(want map[string]string) statecheck.StateCheck {
		return tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*models.ApplicationView, error) {
			appRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Getorganisationapplication(ctx)
			if appRes.HasError() {
				return nil, appRes.Error()
			}
			return appRes.Payload(), nil
		}, func(ctx context.Context, id string, state *tfjson.State, appRes *models.ApplicationView) error {
			appEnvRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(id).Env().Listorganisationapplicationenvironmentvariables(ctx)
			if appEnvRes.HasError() {
				return fmt.Errorf("failed to get application env: %w", appEnvRes.Error())
			}
			env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e models.AddonEnvironmentView) map[string]string {
				acc[e.Name] = e.Value
				return acc
			})
			for name, value := range want {
				if env[name] != value {
					return tests.AssertError("bad env var value "+name, value, env[name])
				}
			}
			return nil
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// First, so the post-test destroy runs with a valid configuration.
			ResourceName: rName,
			Config:       config(map[string]any{"environment": map[string]any{app.CC_GO_BUILD_TOOL: "gomod"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_GO_BUILD_TOOL is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_tool"), knownvalue.StringExact("makefile")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("binary"), knownvalue.StringExact("bin/server")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_GO_BUILD_TOOL: "makefile",
					app.CC_GO_BINARY:     "bin/server",
					app.CC_GO_PKG:        "example.com/app/cmd/server",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"build_tool": "gomod", "run_dir": "example.com/app"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_tool"), knownvalue.StringExact("gomod")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("run_dir"), knownvalue.StringExact("example.com/app")),
				remoteEnv(map[string]string{
					app.CC_GO_BUILD_TOOL: "gomod",
					app.CC_GO_RUNDIR:     "example.com/app",
				}),
			},
		}},
	})
}
