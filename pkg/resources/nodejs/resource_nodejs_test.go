package nodejs_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

//go:embed resource_nodejs_test_block.tf
var nodejsBlock string

//go:embed provider_test_block.tf
var providerBlock string

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

func TestAccNodejs_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-node-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_nodejs.%s", rName)
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
			Config:       fmt.Sprintf(providerBlock, org) + fmt.Sprintf(nodejsBlock, rName, rName),
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
						return fmt.Errorf("invalid name, got: '%s', expect: '%s'", app.Name, rName)
					}

					appEnvRes := tmp.GetAppEnv(ctx, cc, org, id)
					if appEnvRes.HasError() {
						return fmt.Errorf("failed to get application: %w", appEnvRes.Error())
					}

					env := pkg.Reduce(*appEnvRes.Payload(), map[string]string{}, func(acc map[string]string, e tmp.Env) map[string]string {
						acc[e.Name] = e.Value
						fmt.Printf("%+v\n", acc)
						return acc
					})

					v := env["MY_KEY"]
					if v != "myval" {
						return fmt.Errorf("expect env var '%s' set to '%s', but got: '%s'", "MY_KEY", "myval3", v)
					}

					v2 := env["APP_FOLDER"]
					if v2 != "./app" {
						return fmt.Errorf("expect env var '%s' set to '%s', but got: '%s'", "APP_FOLER", "./app", v2)
					}

					return nil
				},
			),
		}},
	})
}
