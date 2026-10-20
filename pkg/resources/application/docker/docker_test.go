package docker_test

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"os"
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

func TestAccDocker_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-docker")
	fullName := fmt.Sprintf("clevercloud_docker.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	dockerBlock := helper.NewRessource(
		"clevercloud_docker",
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
			Config:       providerBlock.Append(dockerBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
			},
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				dockerBlock.SetOneValue("biggest_flavor", "XS"),
			).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("biggest_flavor"), knownvalue.StringExact("XS")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDocker_privateRepository(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-docker")
	fullName := fmt.Sprintf("clevercloud_docker.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	cloneAuth := os.Getenv("CLONE_AUTH")
	dockerBlock := helper.NewRessource(
		"clevercloud_docker",
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
			"repository":           "https://github.com/CleverCloud/some-private-repo.git",
			"commit":               "db165c4e741b450d89eceb931402769e3b7cd7a0",
			"authentication_basic": cloneAuth,
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck: func() {
			if cloneAuth == "" {
				t.Fatalf("missing CLONE_AUTH value")
			}

			tests.ExpectOrganisation(t)
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(dockerBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^app_.*$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("deploy_url"), knownvalue.StringRegexp(regexp.MustCompile(`^git\+ssh.*\.git$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("region"), knownvalue.StringExact("par")),
			},
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDocker_invalidFlavor(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-docker")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	dockerBlock := helper.NewRessource(
		"clevercloud_docker",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":               rName,
			"region":             "par",
			"min_instance_count": 1,
			"max_instance_count": 2,
			"smallest_flavor":    "something",
			"biggest_flavor":     "M",
		}))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(dockerBlock).String(),
			ExpectError:  regexp.MustCompile("invalid flavor"),
		}},
		CheckDestroy: tests.CheckDestroy(ctx),
	})
}

func TestAccDocker_envAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-docker-env")
	fullName := fmt.Sprintf("clevercloud_docker.%s", rName)
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
			"dockerfile":          "Dockerfile.prod",
			"container_port":      3000,
			"registry_url":        "registry.example.com",
			"registry_user":       "user",
			"registry_password":   "secret",
			"daemon_socket_mount": true,
			"buildx":              false,
		}
		maps.Copy(values, overrides)
		return providerBlock.Append(helper.NewRessource("clevercloud_docker", rName, helper.SetKeyValues(values))).String()
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
			Config:       config(map[string]any{"environment": map[string]any{app.CC_DOCKER_EXPOSED_HTTP_PORT: "4000"}}),
			PlanOnly:     true,
			ExpectError:  regexp.MustCompile(`CC_DOCKER_EXPOSED_HTTP_PORT is managed by an attribute`),
		}, {
			ResourceName: rName,
			Config:       config(nil),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("container_port"), knownvalue.Int64Exact(3000)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("buildx"), knownvalue.Bool(false)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("daemon_socket_mount"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("environment"), knownvalue.Null()),
				remoteEnv(map[string]string{
					app.CC_DOCKERFILE:               "Dockerfile.prod",
					app.CC_DOCKER_EXPOSED_HTTP_PORT: "3000",
					"CC_DOCKER_LOGIN_SERVER":        "registry.example.com",
					"CC_DOCKER_LOGIN_USERNAME":      "user",
					"CC_DOCKER_LOGIN_PASSWORD":      "secret",
					app.CC_MOUNT_DOCKER_SOCKET:      "true",
					app.CC_DOCKER_BUILDX:            "false",
				}),
			},
		}, {
			ResourceName: rName,
			Config:       config(map[string]any{"container_port": 4000, "buildx": true}),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("container_port"), knownvalue.Int64Exact(4000)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("buildx"), knownvalue.Bool(true)),
				remoteEnv(map[string]string{
					app.CC_DOCKER_EXPOSED_HTTP_PORT: "4000",
					app.CC_DOCKER_BUILDX:            "true",
				}),
			},
		}},
	})
}
