package statestore_test

import (
	"testing"

	r "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// Pluggable state storage is experimental: it lands in Terraform 1.15
// pre-releases behind a flag. These tests enable it via env var and skip when
// the active CLI is too old or stable.
//
// State store support is gated on:
//   - TF_ENABLE_PLUGGABLE_STATE_STORAGE=1 (set by t.Setenv)
//   - terraform CLI >= 1.15.0 alpha/beta (SkipBelow + SkipIfNotPrerelease)
//   - terraform-plugin-framework v1.19+ (carrier of the statestore package)
var stateStoreVersionChecks = []tfversion.TerraformVersionCheck{
	tfversion.SkipBelow(tfversion.Version1_15_0),
	tfversion.SkipIfNotPrerelease(),
}

const stateStoreConfig = `
terraform {
  required_providers {
    clevercloud = {
      source = "registry.terraform.io/CleverCloud/clevercloud"
    }
  }
  state_store "clevercloud_statestore" {
    provider "clevercloud" {}
  }
}
`

const stateStoreConfigWithLockDuration = `
terraform {
  required_providers {
    clevercloud = {
      source = "registry.terraform.io/CleverCloud/clevercloud"
    }
  }
  state_store "clevercloud_statestore" {
    provider "clevercloud" {}
    lock_duration = "10m"
  }
}
`

// TestAccStateStore_basic exercises the framework's StateStore mode: init,
// workspace creation/listing/deletion, write+read state through real
// Terraform CLI commands.
func TestAccStateStore_basic(t *testing.T) {
	t.Setenv("TF_ENABLE_PLUGGABLE_STATE_STORAGE", "1")

	r.UnitTest(t, r.TestCase{
		TerraformVersionChecks:   stateStoreVersionChecks,
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []r.TestStep{
			{
				StateStore: true,
				Config:     stateStoreConfig,
			},
		},
	})
}

// TestAccStateStore_verifyLock exercises the lock + unlock RPCs through real
// terraform apply commands and asserts that a held lock prevents a second
// acquisition.
func TestAccStateStore_verifyLock(t *testing.T) {
	t.Setenv("TF_ENABLE_PLUGGABLE_STATE_STORAGE", "1")

	r.UnitTest(t, r.TestCase{
		TerraformVersionChecks:   stateStoreVersionChecks,
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []r.TestStep{
			{
				StateStore:           true,
				VerifyStateStoreLock: true,
				Config:               stateStoreConfig,
			},
		},
	})
}

// TestAccStateStore_lockDurationConfig confirms that the lock_duration
// attribute is accepted and parsed end-to-end. Behaviour beyond that
// (TTL-driven lock expiry) is hard to exercise through the CLI without
// time travel, so it stays out of acceptance scope.
func TestAccStateStore_lockDurationConfig(t *testing.T) {
	t.Setenv("TF_ENABLE_PLUGGABLE_STATE_STORAGE", "1")

	r.UnitTest(t, r.TestCase{
		TerraformVersionChecks:   stateStoreVersionChecks,
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []r.TestStep{
			{
				StateStore: true,
				Config:     stateStoreConfigWithLockDuration,
			},
		},
	})
}
