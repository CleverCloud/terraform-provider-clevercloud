package bucket_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccCellarBucket_basic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("my-bucket")
	fullName := "clevercloud_cellar_bucket." + rName
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	cellarBlock := helper.NewRessource(
		"clevercloud_cellar",
		"cellar1",
		helper.SetKeyValues(map[string]any{
			"name": "cellar1",
		}),
	)

	cellarBucketBlock := helper.NewRessource(
		"clevercloud_cellar_bucket",
		rName,
		helper.SetKeyValues(map[string]any{
			"id":        rName,
			"cellar_id": "${clevercloud_cellar.cellar1.id}",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName:      fullName,
			Config:            providerBlock.Append(cellarBlock, cellarBucketBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				// TODO
			},
		}},
		CheckDestroy: func(state *terraform.State) error {
			for resourceName, resourceState := range state.RootModule().Resources {
				switch resourceName {
				case "clevercloud_cellar.cellar1":
					t.Logf("skip cellar addon")

				case fullName:
					id := resourceState.Primary.Attributes["cellar_id"]
					res := tmp.GetAddonEnv(ctx, cc, tests.ORGANISATION, id)
					if res.IsNotFoundError() {
						continue
					}
					if res.HasError() {
						return fmt.Errorf("unexpectd error: %s", res.Error().Error())
					}

					minioClient, err := s3.MinioClientFromEnvsFor(*res.Payload())
					if err != nil {
						return fmt.Errorf("unexpectd error: %s", res.Error().Error())
					}

					exists, err := minioClient.BucketExists(ctx, rName)
					if err != nil {
						return fmt.Errorf("unexpectd error: %s", res.Error().Error())
					}

					if exists {
						return fmt.Errorf("expect cellar bucket resource '%s' to be deleted", resourceName)
					}
				default:
					return fmt.Errorf("unhandled resource: %s", resourceName)
				}
			}
			return nil
		},
	})
}
