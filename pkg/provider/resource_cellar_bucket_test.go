package provider

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

//go:embed resource_cellar_bucket_test_block.tf
var cellarBucketBlock string

func TestAccCellarBucket_basic(t *testing.T) {
	ctx := context.Background()
	bName := fmt.Sprintf("my-bucket-%d", time.Now().UnixMilli())
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

	cellar := &tmp.AddonResponse{}
	if os.Getenv("TF_ACC") == "1" {
		res := tmp.CreateAddon(ctx, cc, org, tmp.AddonRequest{
			Name:       fmt.Sprintf("tf-cellar-%d-forbucket", time.Now().UnixMilli()),
			ProviderID: "cellar-addon",
			Plan:       "plan_84c85ee3-5fdb-4aca-a727-298ddc14b766",
			Region:     "par",
		})
		if res.HasError() {
			t.Errorf("failed to create depdendence Cellar: %s", res.Error().Error())
		}

		cellar = res.Payload()

		defer func() {
			rmRes := tmp.DeleteAddon(ctx, cc, org, cellar.ID)
			if rmRes.HasError() && !rmRes.IsNotFoundError() {
				t.Fatalf("failed to destroy deps %s: %s", cellar.RealID, rmRes.Error().Error())
			}
		}()
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if org == "" {
				t.Fatalf("missing ORGANISATION env var")
			}
			if cellar.RealID == "" {
				t.Fatalf("missing CellarID")
			}
		},
		ProtoV6ProviderFactories: protoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: "cellar_bucket_" + bName,
			Config:       fmt.Sprintf(providerBlock, org) + fmt.Sprintf(cellarBucketBlock, bName, bName, cellar.RealID),
			Check: resource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					return nil
				},
			),
		}},
		CheckDestroy: func(state *terraform.State) error {
			for resourceName, resourceState := range state.RootModule().Resources {
				tflog.Info(ctx, "TEST DESTROY", map[string]interface{}{"bucket": resourceState})
				res := tmp.GetAddonEnv(context.Background(), cc, org, cellar.ID) // TODO: resourceState.Primary.ID)
				if res.IsNotFoundError() {
					continue
				}
				if res.HasError() {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				minioClient, err := minioClientFromEnvsFor(*res.Payload())
				if err != nil {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				exists, err := minioClient.BucketExists(ctx, bName)
				if err != nil {
					return fmt.Errorf("unexpectd error: %s", res.Error().Error())
				}

				if exists {
					return fmt.Errorf("expect cellar bucket resource '%s' to be deleted", resourceName)
				}
			}
			return nil
		},
	})
}
