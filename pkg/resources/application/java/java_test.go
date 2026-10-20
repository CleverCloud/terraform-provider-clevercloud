package java_test

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

func TestAccJava_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-java")
	fullName := fmt.Sprintf("clevercloud_java_war.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	javaBlock := helper.NewRessource(
		"clevercloud_java_war",
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
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			Destroy:      false,
			ResourceName: rName,
			Config:       providerBlock.Append(javaBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				javaBlock.SetOneValue("biggest_flavor", "XS"),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("XS")),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				javaBlock.SetOneValue("java_version", "21"),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("XS")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccJava_jar(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-java")
	fullName := fmt.Sprintf("clevercloud_java_jar.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	javaBlock := helper.NewRessource(
		"clevercloud_java_jar",
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
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(javaBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccJava_maven(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-java")
	fullName := fmt.Sprintf("clevercloud_java_maven.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	javaBlock := helper.NewRessource(
		"clevercloud_java_maven",
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
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(javaBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccJava_empty_vhosts(t *testing.T) {
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-java")
	fullName := fmt.Sprintf("clevercloud_java_war.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	javaBlock := helper.NewRessource(
		"clevercloud_java_war",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "XS",
			"biggest_flavor":     "M",
			"vhosts":             []string{},
		}))

	config := providerBlock.Append(javaBlock).String()

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			Destroy:      false,
			ResourceName: rName,
			Config:       config,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("vhosts"), knownvalue.SetSizeExact(0)),
				tests.NewCheckRemoteResource(fullName, func(ctx context.Context, id string) (*tmp.VHosts, error) {
					appRes := tmp.GetAppVhosts(ctx, cc, tests.ORGANISATION, id)
					if appRes.HasError() {
						return nil, appRes.Error()
					}
					return appRes.Payload(), nil
				}, func(ctx context.Context, id string, state *tfjson.State, vhosts *tmp.VHosts) error {
					if len(*vhosts) != 0 {
						return fmt.Errorf("expect not vhosts, got %d", len(*vhosts))
					}
					return nil
				}),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccJava_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-java-env")
	fullName := fmt.Sprintf("clevercloud_java_jar.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// config builds each step's configuration from scratch: SetOneValue
	// mutates the block, and steps must not leak values into each other.
	config := func(overrides map[string]any) string {
		values := map[string]any{
			"name":                  rName,
			"region":                "par",
			"min_instance_count":    1,
			"max_instance_count":    1,
			"smallest_flavor":       "XS",
			"biggest_flavor":        "XS",
			"jar_path":              "target/app.jar",
			"jar_args":              "--server.port=8080",
			"maven_profiles":        "prod",
			"disable_max_metaspace": true,
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_java_jar", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_JAR_PATH: "other.jar"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_JAR_PATH is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("jar_path"), knownvalue.StringExact("target/app.jar")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("disable_max_metaspace"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_JAR_PATH:              "target/app.jar",
					app.CC_JAR_ARGS:              "--server.port=8080",
					app.CC_MAVEN_PROFILES:        "prod",
					app.CC_DISABLE_MAX_METASPACE: "true",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"jar_path": "build/libs/app.jar", "maven_profiles": "staging"}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("jar_path"), knownvalue.StringExact("build/libs/app.jar")),
				remoteEnv(map[string]string{
					app.CC_JAR_PATH:       "build/libs/app.jar",
					app.CC_MAVEN_PROFILES: "staging",
				}),
			},
		}},
	})
}
