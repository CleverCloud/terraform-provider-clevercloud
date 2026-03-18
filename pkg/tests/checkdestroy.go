package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// CheckDestroy creates a CheckDestroy function that uses the provided context.
// This is a helper function that returns a closure suitable for use in Terraform acceptance tests.
//
// Usage in acceptance tests:
//
//	resource.Test(t, resource.TestCase{
//	    CheckDestroy: tests.CheckDestroy(ctx),
//	    // ... other test configuration
//	})
func CheckDestroy(ctx context.Context) func(*terraform.State) error {
	return func(state *terraform.State) error {
		return checkDestroy(ctx, state)
	}
}

// checkDestroy is the internal implementation that verifies resources are destroyed.
func checkDestroy(ctx context.Context, state *terraform.State) error {
	cc := client.New(client.WithAutoOauthConfig())

	for resourceName, resource := range state.RootModule().Resources {
		// Extract resource type from resource (e.g., "clevercloud_python")
		resourceType := resource.Type
		resourceID := resource.Primary.ID

		var err error
		switch {
		// Applications (all types) - check app is deleted or in TO_DELETE state
		case resourceType == "clevercloud_docker",
			resourceType == "clevercloud_dotnet",
			resourceType == "clevercloud_frankenphp",
			resourceType == "clevercloud_golang",
			resourceType == "clevercloud_java",
			resourceType == "clevercloud_linux",
			resourceType == "clevercloud_nodejs",
			resourceType == "clevercloud_php",
			resourceType == "clevercloud_play2",
			resourceType == "clevercloud_python",
			resourceType == "clevercloud_ruby",
			resourceType == "clevercloud_rust",
			resourceType == "clevercloud_scala",
			resourceType == "clevercloud_static",
			resourceType == "clevercloud_v":
			err = checkApplicationDestroyed(ctx, cc, resourceID, resourceName)

		// Drains - check parent application (drains don't have independent lifecycle)
		// All drain types check if the parent application is deleted
		case strings.HasPrefix(resourceType, "clevercloud_drain_"):
			err = checkApplicationDestroyed(ctx, cc, resourceID, resourceName)

		// Elasticsearch (special case with ID conversion and Status field)
		case resourceType == "clevercloud_elasticsearch":
			err = checkElasticsearchDestroyed(ctx, cc, resourceID, resourceName)

		// ConfigProvider (special case with ID conversion and Status field)
		case resourceType == "clevercloud_configprovider":
			err = checkConfigProviderDestroyed(ctx, cc, resourceID, resourceName)

		// Network Groups
		case resourceType == "clevercloud_networkgroup":
			err = checkNetworkgroupDestroyed(ctx, cc, resourceID, resourceName)

		// All other addons (database, software, storage, etc.)
		case resourceType == "clevercloud_cellar",
			resourceType == "clevercloud_fsbucket",
			resourceType == "clevercloud_keycloak",
			resourceType == "clevercloud_kubernetes",
			resourceType == "clevercloud_materiakv",
			resourceType == "clevercloud_matomo",
			resourceType == "clevercloud_metabase",
			resourceType == "clevercloud_mongodb",
			resourceType == "clevercloud_mysql",
			resourceType == "clevercloud_otoroshi",
			resourceType == "clevercloud_postgresql",
			resourceType == "clevercloud_pulsar",
			resourceType == "clevercloud_redis":
			err = checkAddonDestroyed(ctx, cc, resourceID, resourceName)

		// Skip data sources - they don't need destruction checks
		case strings.HasPrefix(resourceType, "data."):
			continue

		default:
			// Unknown resource type - warn but don't fail the test
			fmt.Printf("Warning: CheckDestroy encountered unknown resource type: %s (resource: %s)\n", resourceType, resourceName)
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// checkApplicationDestroyed verifies that an application has been deleted or is in TO_DELETE state.
// This is also used for drains since they verify the parent application's state.
func checkApplicationDestroyed(ctx context.Context, cc *client.Client, appID, resourceName string) error {
	res := tmp.GetApp(ctx, cc, ORGANISATION, appID)

	if res.IsNotFoundError() {
		return nil // Successfully destroyed
	}

	if res.HasError() {
		return fmt.Errorf("unexpected error checking if application %s was destroyed: %s", resourceName, res.Error().Error())
	}

	// Application exists - check if it's in deletion state
	if res.Payload().State == "TO_DELETE" {
		return nil // In process of being deleted
	}

	return fmt.Errorf("expected application %s (ID: %s) to be deleted but it still exists with state: %s", resourceName, appID, res.Payload().State)
}

// checkAddonDestroyed verifies that an addon has been deleted.
func checkAddonDestroyed(ctx context.Context, cc *client.Client, addonID, resourceName string) error {
	res := tmp.GetAddon(ctx, cc, ORGANISATION, addonID)

	if res.IsNotFoundError() {
		return nil // Successfully destroyed
	}

	if res.HasError() {
		return fmt.Errorf("unexpected error checking if addon %s was destroyed: %s", resourceName, res.Error().Error())
	}

	return fmt.Errorf("expected addon %s (ID: %s) to be deleted but it still exists: %+v", resourceName, addonID, res.Payload())
}

// checkElasticsearchDestroyed verifies that an Elasticsearch addon has been deleted or is in TO_DELETE state.
// Elasticsearch uses a different ID format - need to convert real ID to addon ID first.
func checkElasticsearchDestroyed(ctx context.Context, cc *client.Client, realID, resourceName string) error {
	// Elasticsearch uses a different ID format - need to convert real ID to addon ID
	addonID, err := tmp.RealIDToAddonID(ctx, cc, ORGANISATION, realID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil // Successfully destroyed
		}
		return fmt.Errorf("failed to get addon ID for Elasticsearch %s: %s", resourceName, err.Error())
	}

	res := tmp.GetElasticsearch(ctx, cc, addonID)

	if res.IsNotFoundError() {
		return nil // Successfully destroyed
	}

	if res.HasError() {
		return fmt.Errorf("unexpected error checking if Elasticsearch %s was destroyed: %s", resourceName, res.Error().Error())
	}

	// Elasticsearch exists - check if it's in deletion state
	if res.Payload().Status == "TO_DELETE" {
		return nil // In process of being deleted
	}

	return fmt.Errorf("expected Elasticsearch %s (ID: %s) to be deleted but it still exists", resourceName, realID)
}

// checkConfigProviderDestroyed verifies that a ConfigProvider has been deleted or is in TO_DELETE state.
// ConfigProvider uses the same pattern as Elasticsearch - ID conversion required.
func checkConfigProviderDestroyed(ctx context.Context, cc *client.Client, realID, resourceName string) error {
	// ConfigProvider uses a different ID format - need to convert real ID to addon ID
	addonID, err := tmp.RealIDToAddonID(ctx, cc, ORGANISATION, realID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil // Successfully destroyed
		}
		return fmt.Errorf("failed to get addon ID for ConfigProvider %s: %s", resourceName, err.Error())
	}

	res := tmp.GetConfigProvider(ctx, cc, addonID)

	if res.IsNotFoundError() {
		return nil // Successfully destroyed
	}

	if res.HasError() {
		return fmt.Errorf("unexpected error checking if ConfigProvider %s was destroyed: %s", resourceName, res.Error().Error())
	}

	// ConfigProvider exists - check if it's in deletion state
	if res.Payload().Status == "TO_DELETE" {
		return nil // In process of being deleted
	}

	return fmt.Errorf("expected ConfigProvider %s (ID: %s) to be deleted but it still exists", resourceName, realID)
}

// checkNetworkgroupDestroyed verifies that a network group has been deleted.
func checkNetworkgroupDestroyed(ctx context.Context, cc *client.Client, ngID, resourceName string) error {
	res := tmp.GetNetworkgroup(ctx, cc, ORGANISATION, ngID)

	if res.IsNotFoundError() {
		return nil // Successfully destroyed
	}

	if res.HasError() {
		return fmt.Errorf("unexpected error checking if network group %s was destroyed: %s", resourceName, res.Error().Error())
	}

	return fmt.Errorf("expected network group %s (ID: %s) to be deleted but it still exists", resourceName, ngID)
}
