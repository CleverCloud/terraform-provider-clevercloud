package ruby_test

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"testing"

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

func TestAccRuby_basic(t *testing.T) {
	ctx := context.Background()
	rName := acctest.RandomWithPrefix("tf-test-ruby")
	fullName := fmt.Sprintf("clevercloud_ruby.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
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
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				res := tmp.GetApp(ctx, cc, tests.ORGANISATION, resource.Primary.ID)
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
			Config:       providerBlock.Append(rubyBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("build_flavor"), knownvalue.StringExact("XL")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ruby_version"), knownvalue.StringExact("3.3")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("enable_sidekiq"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("rails_env"), knownvalue.StringExact("production")),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					if app.Name != rName {
						return tests.AssertError("invalid name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return tests.AssertError("invalid min instance count", app.Instance.MinInstances, "1")
					}

					if app.Instance.MaxInstances != 2 {
						return tests.AssertError("invalid max instance count", app.Instance.MaxInstances, 2)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return tests.AssertError("invalid smallest flavor", app.Instance.MinFlavor.Name, "XS")
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return tests.AssertError("invalid biggest flavor", app.Instance.MaxFlavor.Name, "M")
					}

					if app.BuildFlavor.Name != "XL" {
						return tests.AssertError("invalid build flavor", app.BuildFlavor.Name, "XL")
					}

					if app.ForceHTTPS != "ENABLED" {
						return tests.AssertError("expect redirect_https to be set", "redirect_https", app.ForceHTTPS)
					}

					if !app.StickySessions {
						return tests.AssertError("expect sticky_sessions to be set", "sticky_sessions", app.StickySessions)
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

					// Check custom environment variables
					if v := env["MY_KEY"]; v != "myval" {
						return tests.AssertError("bad env var value MY_KEY", "myval", v)
					}

					if v := env["RAILS_SECRET"]; v != "secret123" {
						return tests.AssertError("bad env var value RAILS_SECRET", "secret123", v)
					}

					// Check Ruby-specific environment variables
					if v := env["APP_FOLDER"]; v != "./app" {
						return tests.AssertError("bad env var value APP_FOLDER", "./app", v)
					}

					if v := env["CC_RUBY_VERSION"]; v != "3.3" {
						return tests.AssertError("bad env var value CC_RUBY_VERSION", "3.3", v)
					}

					if v := env["CC_ENABLE_SIDEKIQ"]; v != "true" {
						return tests.AssertError("bad env var value CC_ENABLE_SIDEKIQ", "true", v)
					}

					if v := env["CC_RACKUP_SERVER"]; v != "puma" {
						return tests.AssertError("bad env var value CC_RACKUP_SERVER", "puma", v)
					}

					if v := env["CC_RAKEGOALS"]; v != "db:migrate,assets:precompile" {
						return tests.AssertError("bad env var value CC_RAKEGOALS", "db:migrate,assets:precompile", v)
					}

					if v := env["CC_SIDEKIQ_FILES"]; v != "./config/sidekiq.yml" {
						return tests.AssertError("bad env var value CC_SIDEKIQ_FILES", "./config/sidekiq.yml", v)
					}

					if v := env["RAILS_ENV"]; v != "production" {
						return tests.AssertError("bad env var value RAILS_ENV", "production", v)
					}

					if v := env["RACK_ENV"]; v != "production" {
						return tests.AssertError("bad env var value RACK_ENV", "production", v)
					}

					if v := env["ENABLE_GZIP_COMPRESSION"]; v != "true" {
						return tests.AssertError("bad env var value ENABLE_GZIP_COMPRESSION", "true", v)
					}

					if v := env["NGINX_READ_TIMEOUT"]; v != "600" {
						return tests.AssertError("bad env var value NGINX_READ_TIMEOUT", "600", v)
					}

					if v := env["STATIC_FILES_PATH"]; v != "public" {
						return tests.AssertError("bad env var value STATIC_FILES_PATH", "public", v)
					}

					if v := env["STATIC_URL_PREFIX"]; v != "/assets" {
						return tests.AssertError("bad env var value STATIC_URL_PREFIX", "/assets", v)
					}

					if v := env["CC_POST_BUILD_HOOK"]; v != "bundle exec rake assets:precompile" {
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
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.CreatAppResponse, error) {
					appRes := tmp.GetApp(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, app *tmp.CreatAppResponse) error {
					// Verify instance counts were updated
					if app.Instance.MinInstances != 2 {
						return tests.AssertError("invalid updated min instance count", app.Instance.MinInstances, 2)
					}
					if app.Instance.MaxInstances != 6 {
						return tests.AssertError("invalid updated max instance count", app.Instance.MaxInstances, 6)
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

					if v := env["CC_RUBY_VERSION"]; v != "3.2" {
						return tests.AssertError("bad updated env var value CC_RUBY_VERSION", "3.2", v)
					}

					if v := env["CC_RACKUP_SERVER"]; v != "unicorn" {
						return tests.AssertError("bad updated env var value CC_RACKUP_SERVER", "unicorn", v)
					}

					if v := env["NGINX_READ_TIMEOUT"]; v != "900" {
						return tests.AssertError("bad updated env var value NGINX_READ_TIMEOUT", "900", v)
					}

					// Verify sidekiq is disabled (should not be present or be "false")
					if v := env["CC_ENABLE_SIDEKIQ"]; v != "" && v != "false" {
						return tests.AssertError("expected CC_ENABLE_SIDEKIQ to be disabled", "", v)
					}

					return nil
				}),
			},
		}},
	})
}
