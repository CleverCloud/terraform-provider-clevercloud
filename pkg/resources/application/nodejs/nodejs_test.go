package nodejs_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccNodejs_basic(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-node")
	rName2 := acctest.RandomWithPrefix("tf-test-node-2")
	fullName := fmt.Sprintf("clevercloud_nodejs.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_nodejs.%s", rName2)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	nodejsBlock := helper.NewRessource(
		"clevercloud_nodejs",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"build_flavor":       "XL",
			"redirect_https":     true,
			"sticky_sessions":    true,
			"app_folder":         "./app",
			"environment":        map[string]any{"MY_KEY": "myval"},
			"dependencies":       []string{},
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
	)
	nodejsBlock2 := helper.NewRessource(
		"clevercloud_nodejs",
		rName2,
		helper.SetKeyValues(map[string]any{
			"name":               rName2,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
		}),
		helper.SetBlockValues("deployment", map[string]any{
			"repository": "https://github.com/CleverCloud/nodejs-example.git",
			"commit":     "2474d0e99089096f2e5548e19a2c0ad0f684c674",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(nodejsBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("XL")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.AppResponse) error {
					if app.Name != rName {
						return tests.AssertError("invalid name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return tests.AssertError("invalid min instance count", app.Instance.MinInstances, "1")
					}

					if app.Instance.MaxInstances != 2 {
						return tests.AssertError("invalid name", app.Instance.MaxInstances, 2)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return tests.AssertError("invalid name", app.Instance.MinFlavor.Name, "XS")
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return tests.AssertError("invalid max instance name", app.Instance.MaxFlavor.Name, "M")
					}

					if app.BuildFlavor.Name != "XL" {
						return tests.AssertError("invalid build flavor", app.BuildFlavor.Name, "XL")
					}

					if app.ForceHTTPS != "ENABLED" {
						return tests.AssertError("expect option to be set", "redirect_https", app.ForceHTTPS)
					}

					if !app.StickySessions {
						return tests.AssertError("expect option to be set", "sticky_sessions", app.StickySessions)
					}
					if app.Zone != "par" {
						return tests.AssertError("expect region to be 'par'", "region", app.Zone)
					}
					appEnvRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					v := env["MY_KEY"]
					if v != "myval" {
						return tests.AssertError("bad env var value MY_KEY", "myval3", v)
					}

					v2 := env["APP_FOLDER"]
					if v2 != "./app" {
						return tests.AssertError("bad env var value APP_FOLER", "./app", v2)
					}

					v3 := env["CC_POST_BUILD_HOOK"]
					if v3 != "echo \"build is OK!\"" {
						return tests.AssertError("bad env var value CC_POST_BUILD_HOOK", "echo \"build is OK!\"", v3)
					}
					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				nodejsBlock.SetOneValue("min_instance_count", 2).SetOneValue("max_instance_count", 6),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
			},
		}, {
			ResourceName: rName2,
			Config:       providerBlock.Append(nodejsBlock2).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(fullName2, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.AppResponse) error {
					vhostsRes := tmp.GetAppVhosts(ctx, cc, tests.ORGANISATION, id)
					if vhostsRes.HasError() {
						return fmt.Errorf("failed to get application vhosts: %w", vhostsRes.Error())
					}
					vhosts := vhostsRes.Payload()

					if len(*vhosts) == 0 {
						return fmt.Errorf("there is no vhost for app: %s", id)
					}

					// Test deployed app
					err := tests.HealthCheck(ctx, vhosts.CleverAppsFQDN(id).Fqdn, 2*time.Minute)
					if err != nil {
						return fmt.Errorf("application did not respond in the allowed time: %w", err)
					}

					return nil
				}),
			},
		}},
	})
}

func TestAccNodejs_tcpRedirection(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-node")
	fullName := fmt.Sprintf("clevercloud_nodejs.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	nodejsBlock := helper.NewRessource(
		"clevercloud_nodejs",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 1,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "XS",
			"redirection": map[string]any{
				"namespace": "default",
			},
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			Destroy:      false,
			ResourceName: rName,
			Config:       providerBlock.Append(nodejsBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("redirection").AtMapKey("namespace"), knownvalue.StringExact("default")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("redirection").AtMapKey("port"), knownvalue.NotNull()),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				nodejsBlock.UnsetOneValue("redirection"),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("redirection"), knownvalue.Null()),
			},
		}},
	})
}

// initRepoWithCommit creates a fresh git repository at repoPath holding a
// minimal Node.js HTTP server and returns the hash of its single commit.
func initRepoWithCommit(t *testing.T, repoPath, version string) plumbing.Hash {
	t.Helper()

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	files := map[string]string{
		"package.json": `{"name":"tf-test","version":"1.0.0","scripts":{"start":"node index.js"}}`,
		"index.js": fmt.Sprintf(`const http = require("http");
http.createServer((req, res) => res.end("%s")).listen(process.env.PORT || 8080);
`, version),
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(repoPath, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		if _, err := worktree.Add(name); err != nil {
			t.Fatalf("failed to add %s: %v", name, err)
		}
	}

	hash, err := worktree.Commit("commit "+version, &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	t.Logf("repository %s initialised with commit %s (%s)", repoPath, hash, version)

	return hash
}

// TestAccNodejs_localRepoRecreated covers a local repository deployed without
// an explicit `commit`: the app is created and deploys the repository HEAD
// (c1), then the repository directory is removed and recreated from scratch
// with an unrelated commit (c2). Re-applying the very same configuration must
// deploy c2.
func TestAccNodejs_localRepoRecreated(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-node-repo")
	fullName := fmt.Sprintf("clevercloud_nodejs.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	repoPath := filepath.Join(t.TempDir(), "r1")

	// use a relative path, like `file://../r1` in a real configuration
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	relRepoPath, err := filepath.Rel(cwd, repoPath)
	if err != nil {
		t.Fatalf("failed to build relative repo path: %v", err)
	}
	repoURL := "file://" + relRepoPath
	t.Logf("repository URL: %s", repoURL)

	commit1 := initRepoWithCommit(t, repoPath, "v1")
	var commit2 plumbing.Hash

	config := providerBlock.Append(
		helper.NewRessource(
			"clevercloud_nodejs",
			rName,
			helper.SetKeyValues(map[string]any{
				"name":               rName,
				"region":             "par",
				"min_instance_count": 1,
				"max_instance_count": 1,
				"smallest_flavor":    "XS",
				"biggest_flavor":     "XS",
			}),
			helper.SetBlockValues("deployment", map[string]any{
				"repository": repoURL,
			}),
		),
	).String()

	expectDeployments := func(expectedCount int, expectedCommit *plumbing.Hash) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			rs, ok := state.RootModule().Resources[fullName]
			if !ok {
				return fmt.Errorf("resource not found: %s", fullName)
			}
			appID := rs.Primary.ID

			deployments, err := tests.WaitForDeploymentsSettled(ctx, cc, appID, 10*time.Minute)
			if err != nil {
				return err
			}
			for _, d := range deployments {
				t.Logf("DEPLOY\t%s\t%s\t%s\t%s", d.UUID, d.State, d.Commit, d.Cause)
			}

			if len(deployments) != expectedCount {
				return fmt.Errorf("expected %d deployment(s), got %d", expectedCount, len(deployments))
			}

			latest := deployments[0]
			if latest.Commit != expectedCommit.String() {
				return fmt.Errorf("expected latest deployment to run commit %s, got %s", expectedCommit, latest.Commit)
			}
			if latest.State != "OK" {
				return fmt.Errorf("expected latest deployment to be OK, got %s", latest.State)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			// Step 1: create the app, HEAD (c1) is deployed
			ResourceName: rName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deployment").AtMapKey("commit"), knownvalue.StringExact(commit1.String())),
			},
			Check: expectDeployments(1, &commit1),
		}, {
			// Step 2: the repository is removed and recreated with an
			// unrelated commit (c2), same configuration: c2 must be deployed
			PreConfig: func() {
				if err := os.RemoveAll(repoPath); err != nil {
					t.Fatalf("failed to remove repository: %v", err)
				}
				commit2 = initRepoWithCommit(t, repoPath, "v2")
			},
			ResourceName: rName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				tests.NewCheckRemoteResource(
					fullName,
					func(ctx context.Context, id string) (*tmp.AppResponse, error) {
						appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
						if appRes.HasError() {
							return nil, appRes.Error()
						}
						return appRes.Payload(), nil
					},
					func(ctx context.Context, id string, _ *tfjson.State, app *tmp.AppResponse) error {
						if app.CommitID != commit2.String() {
							return fmt.Errorf("expected application to run commit %s, got %s", commit2, app.CommitID)
						}
						return nil
					},
				),
			},
			Check: expectDeployments(2, &commit2),
		}},
	})
}
