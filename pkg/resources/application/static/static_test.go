package static_test

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
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk/app"
	"go.clever-cloud.dev/sdk/models"
)

func TestAccStatic_basic(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test")
	fullName := fmt.Sprintf("clevercloud_static.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	staticBlock := helper.NewRessource(
		"clevercloud_static",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			Destroy:      false,
			ResourceName: rName,
			Config:       providerBlock.Append(staticBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				staticBlock.SetOneValue("biggest_flavor", "XS"),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("XS")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccStatic_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-static-env")
	fullName := fmt.Sprintf("clevercloud_static.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// config builds each step's configuration from scratch: SetOneValue
	// mutates the block, and steps must not leak values into each other.
	config := func(overrides map[string]any) string {
		values := map[string]any{
			"name":                rName,
			"region":              "par",
			"min_instance_count":  1,
			"max_instance_count":  1,
			"smallest_flavor":     "XS",
			"biggest_flavor":      "XS",
			"webroot":             "public",
			"hugo_version":        "0.140.0",
			"static_server":       "caddy",
			"static_server_flags": "--access-log",
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_static", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_STATIC_SERVER: "static-web-server"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_STATIC_SERVER is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("webroot"), knownvalue.StringExact("public")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("static_server"), knownvalue.StringExact("caddy")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_WEBROOT:       "public",
					app.CC_HUGO_VERSION:  "0.140.0",
					app.CC_STATIC_SERVER: "caddy",
					app.CC_STATIC_FLAGS:  "--access-log",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"webroot": "dist", "static_server": "static-web-server"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("webroot"), knownvalue.StringExact("dist")),
				remoteEnv(map[string]string{
					app.CC_WEBROOT:       "dist",
					app.CC_STATIC_SERVER: "static-web-server",
				}),
			},
		}},
	})
}
