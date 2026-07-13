package nodejs_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"
	"time"

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

func TestAccNodejs_basic(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-node")
	rName2 := acctest.RandomWithPrefix("tf-test-node-2")
	fullName := fmt.Sprintf("clevercloud_nodejs.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_nodejs.%s", rName2)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).
		SetKeyValues(map[string]any{"default_tags": []string{"managed"}})
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
			"tags":               []string{"foo", "bar"},
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
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("foo"),
				})),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("tags_all"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact("bar"),
					knownvalue.StringExact("foo"),
					knownvalue.StringExact("managed"),
				})),
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

// TestAccNodejs_githubHookRequiresBranch checks the shared config validation: a
// GitHub-linked application (deployment.commit = "github_hook") must explicitly
// select the branch to deploy.
func TestAccNodejs_githubHookRequiresBranch(t *testing.T) {
	t.Parallel()

	rName := acctest.RandomWithPrefix("tf-test-node")
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
		}),
		helper.SetBlockValues("deployment", map[string]any{
			"repository": "https://github.com/CleverCloud/nodejs-example.git",
			"commit":     "github_hook",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(nodejsBlock).String(),
			ExpectError:  regexp.MustCompile("branch is required"),
		}},
	})
}

// TestAccNodejs_branchRequiresGithubHook checks the reverse rule: `branch` must
// not be defined on applications not linked to a GitHub repository.
func TestAccNodejs_branchRequiresGithubHook(t *testing.T) {
	t.Parallel()

	rName := acctest.RandomWithPrefix("tf-test-node")
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
			"branch":             "main",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(nodejsBlock).String(),
			ExpectError:  regexp.MustCompile("branch only applies"),
		}},
	})
}
