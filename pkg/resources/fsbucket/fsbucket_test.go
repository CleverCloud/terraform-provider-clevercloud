package fsbucket_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)



func TestAccFSBucket_basic(t *testing.T) {
	ctx := context.Background()
	rName := fmt.Sprintf("tf-test-fsbucket-%d", time.Now().UnixMilli())
	rNameEdited := rName + "-edit"
	fullName := fmt.Sprintf("clevercloud_fsbucket.%s", rName)
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(org)
	fsbucketBlock := helper.NewRessource(
		"clevercloud_fsbucket",
		rName,
		helper.SetKeyValues(map[string]any{"name": rName, "region": "par"}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
		},
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "fsbucket_" + rName,
			Config:       providerBlock.Append(fsbucketBlock).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "name", rName),
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^bucket_.*`)),
				resource.TestMatchResourceAttr(fullName, "host", regexp.MustCompile(`^.*fsbucket.services.clever-cloud.com$`)),
				resource.TestCheckResourceAttrSet(fullName, "ftp_username"),
				resource.TestCheckResourceAttrSet(fullName, "ftp_password"),
			),
		}, {
			ResourceName: "fsbucket_" + rName,
			Config:       providerBlock.Append(fsbucketBlock.SetOneValue("name", rNameEdited)).String(),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(fullName, "name", rNameEdited),
				resource.TestMatchResourceAttr(fullName, "id", regexp.MustCompile(`^bucket_.*`)),
				resource.TestMatchResourceAttr(fullName, "host", regexp.MustCompile(`^.*fsbucket.services.clever-cloud.com$`)),
				resource.TestCheckResourceAttrSet(fullName, "ftp_username"),
				resource.TestCheckResourceAttrSet(fullName, "ftp_password"),
			),
		}},
		CheckDestroy: func(state *terraform.State) error {
			for resourceName, resourceState := range state.RootModule().Resources {
				res := tmp.GetAddon(ctx, cc, org, resourceState.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				return fmt.Errorf("expect fsbucket resource '%s' to be deleted: %+v", resourceName, res.Payload())
			}
			return nil
		},
	})
}
