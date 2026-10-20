package ruby_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"regexp"
	"testing"

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
	"go.clever-cloud.dev/sdk/app"
)

func TestAccRuby_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ruby")
	fullName := fmt.Sprintf("clevercloud_ruby.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	rubyBlock := helper.NewRessource(
		"clevercloud_ruby",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":                    rName,
			"region":                  "par",
			"min_instance_count":      1,
			"max_instance_count":      2,
			"smallest_flavor":         "XS",
			"biggest_flavor":          "M",
			"build_flavor":            "XL",
			"redirect_https":          true,
			"sticky_sessions":         true,
			"app_folder":              "./app",
			"ruby_version":            "3.3",
			"enable_sidekiq":          true,
			"rackup_server":           "puma",
			"rake_goals":              "db:migrate,assets:precompile",
			"sidekiq_files":           "./config/sidekiq.yml",
			"rails_env":               "production",
			"rack_env":                "production",
			"enable_gzip_compression": true,
			"nginx_read_timeout":      600,
			"static_files_path":       "public",
			"static_url_prefix":       "/assets",
			"environment":             map[string]any{"MY_KEY": "myval", "RAILS_SECRET": "secret123"},
			"dependencies":            []string{},
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "bundle exec rake assets:precompile"}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(rubyBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("XL")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ruby_version"), knownvalue.StringExact("3.3")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("enable_sidekiq"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("rails_env"), knownvalue.StringExact("production")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *tmp.AppResponse) error {
					if appRes.Name != rName {
						return tests.AssertError("invalid name", appRes.Name, rName)
					}

					if appRes.Instance.MinInstances != 1 {
						return tests.AssertError("invalid min instance count", appRes.Instance.MinInstances, "1")
					}

					if appRes.Instance.MaxInstances != 2 {
						return tests.AssertError("invalid max instance count", appRes.Instance.MaxInstances, 2)
					}

					if appRes.Instance.MinFlavor.Name != "XS" {
						return tests.AssertError("invalid smallest flavor", appRes.Instance.MinFlavor.Name, "XS")
					}

					if appRes.Instance.MaxFlavor.Name != "M" {
						return tests.AssertError("invalid biggest flavor", appRes.Instance.MaxFlavor.Name, "M")
					}

					if appRes.BuildFlavor.Name != "XL" {
						return tests.AssertError("invalid build flavor", appRes.BuildFlavor.Name, "XL")
					}

					if appRes.ForceHTTPS != "ENABLED" {
						return tests.AssertError("expect redirect_https to be set", "redirect_https", appRes.ForceHTTPS)
					}

					if !appRes.StickySessions {
						return tests.AssertError("expect sticky_sessions to be set", "sticky_sessions", appRes.StickySessions)
					}
					if appRes.Zone != "par" {
						return tests.AssertError("expect region to be 'par'", "region", appRes.Zone)
					}
					appEnvRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					// Check custom environment variables
					if v := env["MY_KEY"]; v != "myval" {
						return tests.AssertError("bad env var value MY_KEY", "myval", v)
					}

					if v := env["RAILS_SECRET"]; v != "secret123" {
						return tests.AssertError("bad env var value RAILS_SECRET", "secret123", v)
					}

					// Check Ruby-specific environment variables
					if v := env[app.APP_FOLDER]; v != "./app" {
						return tests.AssertError("bad env var value APP_FOLDER", "./app", v)
					}

					if v := env[app.CC_RUBY_VERSION]; v != "3.3" {
						return tests.AssertError("bad env var value CC_RUBY_VERSION", "3.3", v)
					}

					if v := env[app.CC_ENABLE_SIDEKIQ]; v != "true" {
						return tests.AssertError("bad env var value CC_ENABLE_SIDEKIQ", "true", v)
					}

					if v := env[app.CC_RACKUP_SERVER]; v != "puma" {
						return tests.AssertError("bad env var value CC_RACKUP_SERVER", "puma", v)
					}

					if v := env[app.CC_RAKEGOALS]; v != "db:migrate,assets:precompile" {
						return tests.AssertError("bad env var value CC_RAKEGOALS", "db:migrate,assets:precompile", v)
					}

					if v := env[app.CC_SIDEKIQ_FILES]; v != "./config/sidekiq.yml" {
						return tests.AssertError("bad env var value CC_SIDEKIQ_FILES", "./config/sidekiq.yml", v)
					}

					if v := env["RAILS_ENV"]; v != "production" {
						return tests.AssertError("bad env var value RAILS_ENV", "production", v)
					}

					if v := env["RACK_ENV"]; v != "production" {
						return tests.AssertError("bad env var value RACK_ENV", "production", v)
					}

					if v := env[app.ENABLE_GZIP_COMPRESSION]; v != "true" {
						return tests.AssertError("bad env var value ENABLE_GZIP_COMPRESSION", "true", v)
					}

					if v := env[app.NGINX_READ_TIMEOUT]; v != "600" {
						return tests.AssertError("bad env var value NGINX_READ_TIMEOUT", "600", v)
					}

					if v := env[app.STATIC_FILES_PATH]; v != "public" {
						return tests.AssertError("bad env var value STATIC_FILES_PATH", "public", v)
					}

					if v := env[app.STATIC_URL_PREFIX]; v != "/assets" {
						return tests.AssertError("bad env var value STATIC_URL_PREFIX", "/assets", v)
					}

					if v := env[app.CC_POST_BUILD_HOOK]; v != "bundle exec rake assets:precompile" {
						return tests.AssertError("bad env var value CC_POST_BUILD_HOOK", "bundle exec rake assets:precompile", v)
					}

					return nil
				}),
			},
		}, {
			// Step 2: Update configuration
			ResourceName: rName,
			Config: providerBlock.Append(
				rubyBlock.
					SetOneValue("min_instance_count", 2).
					SetOneValue("max_instance_count", 6).
					SetOneValue("ruby_version", "3.2").
					SetOneValue("enable_sidekiq", false).
					SetOneValue("rackup_server", "unicorn").
					SetOneValue("nginx_read_timeout", 900).
					SetOneValue("environment", map[string]any{"MY_KEY": "updated_val", "NEW_VAR": "new_value"}),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("min_instance_count"), knownvalue.Int64Exact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("max_instance_count"), knownvalue.Int64Exact(6)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ruby_version"), knownvalue.StringExact("3.2")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("enable_sidekiq"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("rackup_server"), knownvalue.StringExact("unicorn")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("nginx_read_timeout"), knownvalue.Int64Exact(900)),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, appRes *tmp.AppResponse) error {
					// Verify instance counts were updated
					if appRes.Instance.MinInstances != 2 {
						return tests.AssertError("invalid updated min instance count", appRes.Instance.MinInstances, 2)
					}
					if appRes.Instance.MaxInstances != 6 {
						return tests.AssertError("invalid updated max instance count", appRes.Instance.MaxInstances, 6)
					}

					// Verify environment variables were updated
					appEnvRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					if v := env["MY_KEY"]; v != "updated_val" {
						return tests.AssertError("bad updated env var value MY_KEY", "updated_val", v)
					}

					if v := env["NEW_VAR"]; v != "new_value" {
						return tests.AssertError("bad env var value NEW_VAR", "new_value", v)
					}

					if v := env[app.CC_RUBY_VERSION]; v != "3.2" {
						return tests.AssertError("bad updated env var value CC_RUBY_VERSION", "3.2", v)
					}

					if v := env[app.CC_RACKUP_SERVER]; v != "unicorn" {
						return tests.AssertError("bad updated env var value CC_RACKUP_SERVER", "unicorn", v)
					}

					if v := env[app.NGINX_READ_TIMEOUT]; v != "900" {
						return tests.AssertError("bad updated env var value NGINX_READ_TIMEOUT", "900", v)
					}

					// Verify sidekiq is disabled (should not be present or be "false")
					if v := env[app.CC_ENABLE_SIDEKIQ]; v != "" && v != "false" {
						return tests.AssertError("expected CC_ENABLE_SIDEKIQ to be disabled", "", v)
					}

					return nil
				}),
			},
		}},
	})
}

func TestAccRuby_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ruby-env")
	fullName := fmt.Sprintf("clevercloud_ruby.%s", rName)
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
			"ruby_version":       "3.3",
			"rake_goals":         "db:migrate",
			"rails_env":          "production",
			"nginx_read_timeout": 600,
			"enable_sidekiq":     true,
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_ruby", rName, helper.SetKeyValues(values))).String()
	}

	remoteEnv := func(want map[string]string) statecheck.StateCheck {
		return tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.AppResponse, error) {
			appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
			if appRes.HasError() {
				return nil, appRes.Error()
			}
			return appRes.Payload(), nil
		}, func(ctx context.Context, id string, state *tfjson.State, appRes *tmp.AppResponse) error {
			appEnvRes := tmp.GetAppEnv(ctx, cc, tests.ORGANISATION, id)
			if appEnvRes.HasError() {
				return fmt.Errorf("failed to get application env: %w", appEnvRes.Error())
			}
			env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_RUBY_VERSION: "3.4"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_RUBY_VERSION is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ruby_version"), knownvalue.StringExact("3.3")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("nginx_read_timeout"), knownvalue.Int64Exact(600)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("enable_sidekiq"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_RUBY_VERSION:    "3.3",
					app.CC_RAKEGOALS:       "db:migrate",
					"RAILS_ENV":            "production",
					app.NGINX_READ_TIMEOUT: "600",
					app.CC_ENABLE_SIDEKIQ:  "true",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"ruby_version": "3.4", "nginx_read_timeout": 120}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ruby_version"), knownvalue.StringExact("3.4")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("nginx_read_timeout"), knownvalue.Int64Exact(120)),
				remoteEnv(map[string]string{
					app.CC_RUBY_VERSION:    "3.4",
					app.NGINX_READ_TIMEOUT: "120",
				}),
			},
		}},
	})
}
