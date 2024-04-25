package python_test

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"regexp"
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
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

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
			Config: helper.NewProvider("clevercloud").
				SetOrganisation(org).String() + helper.NewRessource(
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
			).String(),
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
						return assertError("invalid name", app.Name, rName)
					}

					if app.Instance.MinInstances != 1 {
						return assertError("invalid min instance count", app.Instance.MinInstances, "1")
					}

					if app.Instance.MaxInstances != 2 {
						return assertError("invalid name", app.Name, rName)
					}

					if app.Instance.MinFlavor.Name != "XS" {
						return assertError("invalid name", app.Name, rName)
					}

					if app.Instance.MaxFlavor.Name != "M" {
						return assertError("invalid name", app.Name, rName)
					}

					if app.ForceHTTPS != "ENABLED" {
						return assertError("expect option to be set", "redirect_https", app.ForceHTTPS)
					}

					if !app.StickySessions {
						return assertError("expect option to be set", "sticky_sessions", app.StickySessions)
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
						return assertError("bad env var value MY_KEY", "myval", v)
					}

					v2 := env["APP_FOLDER"]
					if v2 != "./app" {
						return assertError("bad env var value APP_FOLER", "./app", v2)
					}

					v3 := env["CC_POST_BUILD_HOOK"]
					if v3 != "echo \"build is OK!\"" {
						return assertError("bad env var value CC_POST_BUILD_HOOK", "echo \"build is OK!\"", v3)
					}

					v4 := env["CC_PIP_REQUIREMENTS_FILE"]
					if v4 != "requirements.txt" {
						return assertError("bad env var value CC_PIP_REQUIREMENTS_FILE", "requirements.txt", v4)
					}

					return nil
				},
			),
		}, {
			ResourceName: rName2,
			Config: helper.NewProvider("clevercloud").SetOrganisation(org).String() + helper.NewRessource("clevercloud_python",
				rName2,
				helper.SetKeyValues(map[string]any{
					"name":               rName,
					"region":             "par",
					"min_instance_count": 1,
					"max_instance_count": 2,
					"smallest_flavor":    "XS",
					"biggest_flavor":     "M",
				}),
				helper.SetBlockValues("deployment", map[string]any{"repository": "https://github.com/CleverCloud/flask-example.git"})).String(),
			Check: func(state *terraform.State) error {
				id := state.RootModule().Resources[fullName2].Primary.ID

				appRes := tmp.GetApp(ctx, cc, org, id)
				if appRes.HasError() {
					return fmt.Errorf("failed to get application: %w", appRes.Error())
				}
				app := appRes.Payload()

				// Test deployed app
				t := time.NewTimer(2 * time.Minute)
				select {
				case <-healthCheck(app.Vhosts[0].Fqdn):
					return nil
				case <-t.C:
					return fmt.Errorf("application did not respond in the allowed time")
				}
			},
		}},
	})
}

func assertError(msg string, a, b interface{}) error {
	return fmt.Errorf("%s, got: '%v', expect: '%v'", msg, a, b)
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
