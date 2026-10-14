package bucket_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/minio/minio-go/v7"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/resources/database/cellar/bucket"
	"go.clever-cloud.com/terraform-provider/pkg/s3"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

const (
	testCellarID = "cellar_f014d939-a9f5-49fb-af22-041f49a03cee"
	testAddonID  = "addon_f014d939-a9f5-49fb-af22-041f49a03cee"
)

func TestCellarBucket_importStateID(t *testing.T) {
	cases := []struct {
		name         string
		id           string
		wantErr      bool
		wantName     string
		wantCellarID string
	}{
		{name: "cellar and bucket", id: testCellarID + "/my-bucket", wantName: "my-bucket", wantCellarID: testCellarID},
		{name: "bucket name only", id: "my-bucket", wantErr: true},
		{name: "empty bucket name", id: testCellarID + "/", wantErr: true},
		{name: "empty cellar id", id: "/my-bucket", wantErr: true},
		{name: "add-on id", id: testAddonID + "/my-bucket", wantName: "my-bucket", wantCellarID: testAddonID},
		{name: "other add-on kind", id: "redis_f014d939-a9f5-49fb-af22-041f49a03cee/my-bucket", wantErr: true},
		{name: "slash in bucket name", id: testCellarID + "/my/bucket", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := bucket.NewResourceCellarBucket().(resource.ResourceWithImportState)
			var schema resource.SchemaResponse
			r.Schema(t.Context(), resource.SchemaRequest{}, &schema)
			resp := resource.ImportStateResponse{State: tfsdk.State{Schema: schema.Schema}}
			if diags := resp.State.Set(t.Context(), bucket.CellarBucket{
				Name:     types.StringNull(),
				CellarID: types.StringNull(),
			}); diags.HasError() {
				t.Fatalf("init state: %v", diags)
			}

			r.ImportState(t.Context(), resource.ImportStateRequest{ID: tc.id}, &resp)

			if tc.wantErr {
				if !resp.Diagnostics.HasError() {
					t.Fatalf("ImportState(%q): want error, got none", tc.id)
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("ImportState(%q): %v", tc.id, resp.Diagnostics)
			}
			var got bucket.CellarBucket
			if diags := resp.State.Get(t.Context(), &got); diags.HasError() {
				t.Fatalf("read state: %v", diags)
			}
			if got.Name.ValueString() != tc.wantName || got.CellarID.ValueString() != tc.wantCellarID {
				t.Fatalf("ImportState(%q) = {id: %q, cellar_id: %q}, want {id: %q, cellar_id: %q}",
					tc.id, got.Name.ValueString(), got.CellarID.ValueString(), tc.wantName, tc.wantCellarID)
			}
		})
	}
}

func cellarAndBucket(cellarName, bucketName string) (fmt.Stringer, fmt.Stringer) {
	cellar := helper.NewRessource("clevercloud_cellar", "cellar1",
		helper.SetKeyValues(map[string]any{"name": cellarName}))
	b := helper.NewRessource("clevercloud_cellar_bucket", bucketName,
		helper.SetKeyValues(map[string]any{
			"id":        bucketName,
			"cellar_id": "${clevercloud_cellar.cellar1.id}",
		}))
	return cellar, b
}

func cellarIDFrom(state *terraform.State) string {
	return state.RootModule().Resources["clevercloud_cellar.cellar1"].Primary.ID
}

func s3ClientFor(ctx context.Context, cc *client.Client, cellarID string) (*minio.Client, error) {
	res := tmp.GetAddonEnv(ctx, cc, tests.ORGANISATION, cellarID)
	if res.HasError() {
		return nil, res.Error()
	}
	return s3.MinioClientFromEnvsFor(*res.Payload())
}

func TestAccCellarBucket_import(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-bucket")
	cellar, b := cellarAndBucket(acctest.RandomWithPrefix("tf-test-cellar"), rName)
	config := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).Append(cellar, b).String()

	tfresource.Test(t, tfresource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []tfresource.TestStep{{
			Config: config,
		}, {
			ResourceName: "clevercloud_cellar_bucket." + rName,
			Config:       config,
			ImportState:  true,
			ImportStateIdFunc: func(state *terraform.State) (string, error) {
				return cellarIDFrom(state) + "/" + rName, nil
			},
			ImportStateVerify: true,
		}},
	})
}

func TestAccCellarBucket_importNonExistent(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-bucket")
	cellar, b := cellarAndBucket(acctest.RandomWithPrefix("tf-test-cellar"), rName)
	config := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).Append(cellar, b).String()

	tfresource.Test(t, tfresource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []tfresource.TestStep{{
			Config: config,
		}, {
			ResourceName: "clevercloud_cellar_bucket." + rName,
			Config:       config,
			ImportState:  true,
			ImportStateIdFunc: func(state *terraform.State) (string, error) {
				return cellarIDFrom(state) + "/tf-test-bucket-does-not-exist", nil
			},
			ExpectError: regexp.MustCompile(`Cannot import non-existent remote object`),
		}},
	})
}

// A bucket created outside of Terraform is adopted without any change planned.
func TestAccCellarBucket_importExisting(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-bucket")
	cellar, b := cellarAndBucket(acctest.RandomWithPrefix("tf-test-cellar"), rName)
	provider := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	cellarOnly := provider.Append(cellar).String()
	withBucket := provider.Append(cellar, b).String()

	tfresource.Test(t, tfresource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []tfresource.TestStep{{
			Config: cellarOnly,
			Check: func(state *terraform.State) error {
				mc, err := s3ClientFor(ctx, cc, cellarIDFrom(state))
				if err != nil {
					return err
				}
				return mc.MakeBucket(ctx, rName, minio.MakeBucketOptions{})
			},
		}, {
			ResourceName:       "clevercloud_cellar_bucket." + rName,
			Config:             withBucket,
			ImportState:        true,
			ImportStatePersist: true,
			ImportStateIdFunc: func(state *terraform.State) (string, error) {
				return cellarIDFrom(state) + "/" + rName, nil
			},
		}, {
			Config:   withBucket,
			PlanOnly: true,
		}},
	})
}

// A bucket deleted outside of Terraform is planned for re-creation.
func TestAccCellarBucket_deletedOutOfBand(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-bucket")
	cellar, b := cellarAndBucket(acctest.RandomWithPrefix("tf-test-cellar"), rName)
	config := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).Append(cellar, b).String()

	tfresource.Test(t, tfresource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []tfresource.TestStep{{
			Config: config,
			Check: func(state *terraform.State) error {
				mc, err := s3ClientFor(ctx, cc, cellarIDFrom(state))
				if err != nil {
					return err
				}
				return mc.RemoveBucket(ctx, rName)
			},
			ExpectNonEmptyPlan: true,
		}},
	})
}

// A bucket whose Cellar was deleted outside of Terraform is planned for
// re-creation, it does not make the plan fail.
func TestAccCellarBucket_cellarDeletedOutOfBand(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-bucket")
	cellar, b := cellarAndBucket(acctest.RandomWithPrefix("tf-test-cellar"), rName)
	config := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION).Append(cellar, b).String()

	tfresource.Test(t, tfresource.TestCase{
		PreCheck:                 tests.ExpectOrganisation(t),
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []tfresource.TestStep{{
			Config: config,
			Check: func(state *terraform.State) error {
				addonID, err := tmp.RealIDToAddonID(ctx, cc, tests.ORGANISATION, cellarIDFrom(state))
				if err != nil {
					return err
				}
				if res := tmp.DeleteAddon(ctx, cc, tests.ORGANISATION, addonID); res.HasError() {
					return res.Error()
				}
				return nil
			},
			ExpectNonEmptyPlan: true,
		}, {
			// Refresh drops both resources from state, nothing left to destroy
			RefreshState:       true,
			ExpectNonEmptyPlan: true,
		}},
	})
}
