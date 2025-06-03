package python_test

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

func TestAccPython_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-python-%d", time.Now().UnixMilli())
	rName2 := fmt.Sprintf("tf-test-python-%d-2", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_python.%s", rName)
	fullName2 := fmt.Sprintf("clevercloud_python.%s", rName2)
	vhost := "bubhbfbnriubielrbeuvieuv.com"
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
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
			"app_folder":         "./app",
			"python_version":     "2.7",
			"pip_requirements":   "requirements.txt",
			"environment": map[string]any{
				"MY_KEY": "myval",
			},
		}),
		helper.SetBlockValues("hooks", map[string]any{"post_build": "echo \"build is OK!\""}),
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
				resource.TestMatchResourceAttr(fullName, "deploy_url", regexp.MustCompile(`^git\+ssh.*\.git$`)),
				resource.TestCheckResourceAttr(fullName, "region", "par"),
				// Test CleverCloud API for configured applications
				func(state *terraform.State) error {
					id := state.RootModule().Resources[fullName].Primary.ID

					appRes := tmp.GetApp(ctx, cc, org, id)
					if appRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appRes.Error())
					}
					app := appRes.Payload()

					if app.Name != rName {
						return assertError("invalid name", "name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return assertError("invalid min instance count", "min_instance_count", app.Instance.MinInstances, 1)
					}

					if app.Instance.MaxInstances != 2 {
						return assertError("invalid max instance count", "max_instance_count", app.Instance.MaxInstances, 2)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return assertError("invalid min flavor", "min_flavor", app.Instance.MinFlavor.Name, "XS")
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return assertError("invalid max flavor", "max_flavor", app.Instance.MaxFlavor.Name, "M")
					}

					if app.ForceHTTPS != "ENABLED" {
						return assertError("expect option to be set", "redirect_https", app.ForceHTTPS, "ENABLED")
					}

					if !app.StickySessions {
						return assertError("expect option to be set", "sticky_sessions", app.StickySessions, true)
					}

					if len(app.Vhosts) != 1 || !strings.HasSuffix(app.Vhosts[0].Fqdn, ".cleverapps.io") {
						return assertError("invalid vhost list", "vhosts", app.Vhosts.AsString(), "1 cleverapps.io domain")
					}

					appEnvRes := tmp.GetAppEnv(ctx, cc, org, id)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						return acc
					})

					v := env["MY_KEY"]
					if v != "myval" {
						return assertError("bad env var value", "env:MY_KEY", v, "myval")
					}

					v2 := env["APP_FOLDER"]
					if v2 != "./app" {
						return assertError("bad env var value", "env:APP_FOLDER", v2, "./app")
					}

					v3 := env["CC_POST_BUILD_HOOK"]
					if v3 != "echo \"build is OK!\"" {
						return assertError("bad env var value", "env:CC_POST_BUILD_HOOK", v3, "echo \"build is OK!\"")
					}

					v4 := env["CC_PIP_REQUIREMENTS_FILE"]
					if v4 != "requirements.txt" {
						return assertError("bad env var value", "env:CC_PIP_REQUIREMENTS_FILE", v4, "requirements.txt")
					}

					return nil
				},
			),
		}, {
			ResourceName: rName,
			Config: providerBlock.Append(
				pythonBlock.
					SetOneValue("biggest_flavor", "XS").
					SetOneValue("vhosts", []string{"bubhbfbnriubielrbeuvieuv.com"}),
			).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "biggest_flavor", "XS"),
				func(s *terraform.State) error {
					id := s.RootModule().Resources[fullName].Primary.ID

					appRes := tmp.GetApp(ctx, cc, org, id)
					if appRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appRes.Error())
					}
					app := appRes.Payload()

					if len(app.Vhosts) != 1 || app.Vhosts[0].Fqdn != vhost {
						return assertError("invalid vhost list", "vhosts", app.Vhosts.AsString(), vhost)
					}
					return nil
				},
			),
		}, {
			ResourceName: rName2,
			Config: providerBlock.Append(
				helper.NewRessource(
					"clevercloud_python",
					rName2,
					helper.SetKeyValues(map[string]any{
						"name":               rName,
						"region":             "par",
						"min_instance_count": 1,
						"max_instance_count": 2,
						"smallest_flavor":    "XS",
						"biggest_flavor":     "M",
					}),
					helper.SetBlockValues(
						"deployment",
						map[string]any{"repository": "https://github.com/CleverCloud/flask-example.git"},
					),
				),
			).String(),
			Check: func(state *terraform.State) error {
				id := state.RootModule().Resources[fullName2].Primary.ID

				vhostsRes := tmp.GetAppVhosts(ctx, cc, org, id)
				if vhostsRes.HasError() {
					return fmt.Errorf("failed to get application: %w", vhostsRes.Error())
				}
				vhosts := vhostsRes.Payload()

				if len(*vhosts) == 0 {
					return fmt.Errorf("there is no vhost for app: %s", id)
				}

				// Test deployed app
				t := time.NewTimer(2 * time.Minute)
				select {
				case <-healthCheck(vhosts.CleverAppsFQDN(id).Fqdn):
					return nil
				case <-t.C:
					return fmt.Errorf("application did not respond in the allowed time")
				}
			},
		}},
	})
}

func assertError(msg, param string, got, expect any) error {
	return fmt.Errorf("%s, %s = '%v', expect: '%v'", msg, param, got, expect)
}

func healthCheck(vhost string) chan struct{} {
	c := make(chan struct{})

	fmt.Printf("Test on %s\n", vhost)

	go func() {
		for {
			res, err := http.Get(fmt.Sprintf("https://%s", vhost))
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}

			fmt.Printf("RESPONSE %d\n", res.StatusCode)
			if res.StatusCode != 200 {
				time.Sleep(1 * time.Second)
				continue
			}
			c <- struct{}{}
		}
	}()

	return c
}
