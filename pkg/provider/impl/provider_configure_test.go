package impl_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
)

// envBackup handles backup and restoration of environment variables for tests
func envBackup(vars ...string) func() {
	backup := make(map[string]string)
	for _, v := range vars {
		backup[v] = os.Getenv(v)
	}
	return func() {
		for _, v := range vars {
			if envKey, exists := backup[v]; exists && envKey != "" {
				os.Setenv(v, envKey)
			} else {
				os.Unsetenv(v)
			}
		}
	}
}

// fileBackup handles backup and restoration of files for tests
func fileBackup(t *testing.T, filePath string) func() {
	backupPath := filePath + ".backup"
	var fileExisted bool

	// Backup original file content if it exists
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Rename(filePath, backupPath); err != nil {
			t.Errorf("Failed to create backup file %s: %v", backupPath, err)
		}
		fileExisted = true
	}

	return func() {
		if fileExisted {
			// Ensure directory exists
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Errorf("Failed to create directory %s: %v", dir, err)
				return
			}
			// Restore original file from backup
			if err := os.Rename(backupPath, filePath); err != nil {
				t.Errorf("Failed to restore file %s from backup: %v", filePath, err)
			}
		} else {
			// Remove file if it didn't exist originally
			os.Remove(filePath)
			// Clean up any backup file that might exist
			os.Remove(backupPath)
		}
	}
}

func TestProvider_ConfigureEmptyOrganisation(t *testing.T) {
	providerBlock := `
provider "clevercloud" {}

// dummy resource to trigger provider configuration
resource "clevercloud_cellar" "test" {
  name = "test"
}
`
	expectedError := regexp.MustCompile("Organisation should be set")

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			Config: providerBlock,
			SkipFunc: func() (bool, error) {
				return false, nil
			},
			ExpectError: expectedError,
		}},
	})
}

func TestProvider_ConfigureInvalidEnvironmentVariables(t *testing.T) {
	defer envBackup("CC_OAUTH_TOKEN", "CC_OAUTH_SECRET")()

	// Set invalid credentials and valid organisation
	os.Setenv("CC_OAUTH_TOKEN", "invalid_token")
	os.Setenv("CC_OAUTH_SECRET", "invalid_secret")

	// Create provider and resource using helpers with valid organisation
	provider := helper.NewProvider("clevercloud").SetOrganisation("orga_00000000-0000-0000-0000-000000000000")
	cellar := helper.NewRessource("clevercloud_cellar", "test",
		helper.SetKeyValues(map[string]any{"name": "test"}))

	config := provider.Append(cellar).String()
	expectedError := regexp.MustCompile(`Status 401`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: expectedError,
		}},
	})
}

// TestProvider_ConfigureNoCredentials must not run in parallel with other tests
// as it temporarily removes the clever-tools.json file which affects other tests.
// When running tests that include this one, use: go test -p 1 -run "TestProvider_Configure"
func TestProvider_ConfigureNoCredentials(t *testing.T) {
	defer envBackup("CC_OAUTH_TOKEN", "CC_OAUTH_SECRET")()

	// Backup clever-tools config file
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "clever-cloud", "clever-tools.json")
	defer fileBackup(t, configPath)()

	// Remove clever-tools config to ensure "No credentials found" error
	os.Remove(configPath)

	// Clear all environment credentials
	os.Unsetenv("CC_OAUTH_TOKEN")
	os.Unsetenv("CC_OAUTH_SECRET")

	// Create provider and resource using helpers with valid organisation
	provider := helper.NewProvider("clevercloud").SetOrganisation("orga_00000000-0000-0000-0000-000000000000")
	cellar := helper.NewRessource("clevercloud_cellar", "test",
		helper.SetKeyValues(map[string]any{"name": "test"}))

	config := provider.Append(cellar).String()
	expectedError := regexp.MustCompile(`CleverCloud authentication empty`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: expectedError,
		}},
	})
}
