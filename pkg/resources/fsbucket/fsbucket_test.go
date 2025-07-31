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
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^bucket_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*fsbucket.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ftp_username"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ftp_password"), knownvalue.NotNull()),
			},
		}, {
			ResourceName: "fsbucket_" + rName,
			Config:       providerBlock.Append(fsbucketBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^bucket_.*`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("host"), knownvalue.StringRegexp(regexp.MustCompile(`^.*fsbucket.services.clever-cloud.com$`))),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ftp_username"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("ftp_password"), knownvalue.NotNull()),
			},
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
