package bucket_test

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/minio/minio-go/v7"
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

// TestAccCellarBucket_deleteNonEmpty tests the scenario described in issue #295
// where a bucket with objects cannot be deleted
func TestAccCellarBucket_deleteNonEmpty(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-bucket")
	fullName := "clevercloud_cellar_bucket." + rName
	cc := client.New(client.WithAutoOauthConfig())
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	cellarBlock := helper.NewRessource(
		"clevercloud_cellar",
		"cellar_nonempty",
		helper.SetKeyValues(map[string]any{
			"name": "cellar-nonempty-test",
		}),
	)

	cellarBucketBlock := helper.NewRessource(
		"clevercloud_cellar_bucket",
		rName,
		helper.SetKeyValues(map[string]any{
			"id":        rName,
			"cellar_id": "${clevercloud_cellar.cellar_nonempty.id}",
		}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			ResourceName: fullName,
			Config:       providerBlock.Append(cellarBlock, cellarBucketBlock).String(),
			Check: func(state *terraform.State) error {
				resourceState, ok := state.RootModule().Resources[fullName]
				if !ok {
					return fmt.Errorf("resource %s not found in state", fullName)
				}

				cellarID := resourceState.Primary.Attributes["cellar_id"]
				bucketName := resourceState.Primary.Attributes["id"]

				res := tmp.GetAddonEnv(ctx, cc, tests.ORGANISATION, cellarID)
				if res.HasError() {
					return fmt.Errorf("failed to get cellar env: %s", res.Error().Error())
				}

				minioClient, err := s3.MinioClientFromEnvsFor(*res.Payload())
				if err != nil {
					return fmt.Errorf("failed to create S3 client: %s", err.Error())
				}

				objectName := "test-object.txt"
				objectContent := "This is test content to make the bucket non-empty"
				_, err = minioClient.PutObject(
					ctx,
					bucketName,
					objectName,
					strings.NewReader(objectContent),
					int64(len(objectContent)),
					minio.PutObjectOptions{
						ContentType: "text/plain",
					},
				)
				if err != nil {
					return fmt.Errorf("failed to upload test object: %s", err.Error())
				}

				t.Logf("Successfully uploaded object '%s' to bucket '%s'", objectName, bucketName)
				return nil
			},
		}},
		CheckDestroy: func(state *terraform.State) error {
			// This test is expected to fail with "The bucket you tried to delete is not empty"
			// The CheckDestroy will verify that the bucket still exists (deletion failed)
			for resourceName, resourceState := range state.RootModule().Resources {
				switch resourceName {
				case "clevercloud_cellar.cellar_nonempty":
					t.Logf("skip cellar addon")

				case fullName:
					id := resourceState.Primary.Attributes["cellar_id"]
					bucketName := resourceState.Primary.Attributes["id"]

					res := tmp.GetAddonEnv(ctx, cc, tests.ORGANISATION, id)
					if res.IsNotFoundError() {
						continue
					}
					if res.HasError() {
						return fmt.Errorf("unexpected error: %s", res.Error().Error())
					}

					minioClient, err := s3.MinioClientFromEnvsFor(*res.Payload())
					if err != nil {
						return fmt.Errorf("unexpected error creating client: %s", err.Error())
					}

					// Check if bucket still exists (it should because deletion should have failed)
					exists, err := minioClient.BucketExists(ctx, bucketName)
					if err != nil {
						return fmt.Errorf("unexpected error checking bucket: %s", err.Error())
					}

					if exists {
						t.Logf("Bucket '%s' still exists after destroy attempt (expected due to non-empty bucket)", bucketName)
						// Clean up the bucket manually for the test
						objectsCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
							Recursive: true,
						})
						for object := range objectsCh {
							if object.Err != nil {
								t.Logf("Error listing objects: %s", object.Err)
								continue
							}
							err := minioClient.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
							if err != nil {
								t.Logf("Error removing object %s: %s", object.Key, err)
							}
						}
						// Now remove the empty bucket
						err = minioClient.RemoveBucket(ctx, bucketName)
						if err != nil {
							t.Logf("Error removing bucket after cleanup: %s", err)
						}
					}
				default:
					// Ignore other resources
				}
			}
			return nil
		},
	})
}
