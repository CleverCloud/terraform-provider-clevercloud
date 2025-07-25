package mongodb_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/provider/impl"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

var protoV6Provider = map[string]func() (tfprotov6.ProviderServer, error){
	"clevercloud": providerserver.NewProtocol6WithError(impl.New("test")()),
}

func TestAccMongoDB_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-mg-%d", time.Now().UnixMilli())
	fullName := fmt.Sprintf("clevercloud_mongodb.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	mongodbBlock := helper.NewRessource("clevercloud_mongodb", rName, helper.SetKeyValues(map[string]any{"name": rName, "plan": "xs_med", "region": "par"}))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				addonId, err := tmp.RealIDToAddonID(context.Background(), cc, org, resource.Primary.ID)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						continue
					}
					return fmt.Errorf("failed to get addon ID: %s", err.Error())
				}

				res := tmp.GetMongoDB(context.Background(), cc, addonId)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}
				if res.Payload().Status == "TO_DELETE" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted", resource.Primary.ID)
			}
			return nil
		},
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(mongodbBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^mongodb_.*`)),
			),
		}},
	})
}

func TestAccMongoDB_RefreshDeleted(t *testing.T) {
	rName := fmt.Sprintf("tf-test-mg-%d", time.Now().UnixMilli())
	//fullName := fmt.Sprintf("clevercloud_mongodb.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	mongodbBlock2 := helper.NewRessource("clevercloud_mongodb", rName, helper.SetKeyValues(map[string]any{"name": rName, "plan": "xs_med", "region": "par"}))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		CheckDestroy: func(state *terraform.State) error {
			for _, resource := range state.RootModule().Resources {
				addonId, err := tmp.RealIDToAddonID(context.Background(), cc, org, resource.Primary.ID)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						continue
					}
					return fmt.Errorf("failed to get addon ID: %s", err.Error())
				}

				res := tmp.GetMongoDB(context.Background(), cc, addonId)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}
				if res.Payload().Status == "TO_DELETE" {
					continue
				}

				return fmt.Errorf("expect resource '%s' to be deleted", resource.Primary.ID)
			}
			return nil
		},
		Steps: []resource.TestStep{
			// create a database instance on first step
			{
				ResourceName: rName,
				Config:       providerBlock.Append(mongodbBlock2).String(),
			},
			{
				ResourceName: rName,
				PreConfig: func() {
					// delete the database using an api call
					tmp.DeleteAddon(context.Background(), cc, org, rName)
				},
				// refreshing state
				RefreshState: true,
				// plan should contain database re-creation
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
