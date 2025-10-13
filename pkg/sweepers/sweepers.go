package sweepers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// sweepNetworkgroups removes all test networkgroups
func SweepNetworkgroups(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

	if org == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping networkgroups in organization: %s", org)

	// List all networkgroups
	ngsRes := tmp.ListNetworkgroups(ctx, cc, org)
	if ngsRes.HasError() {
		return fmt.Errorf("failed to list networkgroups: %w", ngsRes.Error())
	}

	ngs := *ngsRes.Payload()
	swept := 0
	errors := 0

	for _, ng := range ngs {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(ng.Label, "tf-test") {
			continue
		}

		log.Printf("[INFO] Deleting networkgroup: %s (%s)", ng.Label, ng.ID)
		delRes := tmp.DeleteNetworkgroup(ctx, cc, org, ng.ID)
		if delRes.HasError() {
			log.Printf("[ERROR] Failed to delete networkgroup %s: %v", ng.ID, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d networkgroups (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping networkgroups", errors)
	}
	return nil
}

// sweepApplications removes all test applications
func SweepApplications(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

	if org == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping applications in organization: %s", org)

	// List all applications
	appsRes := tmp.ListApps(ctx, cc, org)
	if appsRes.HasError() {
		return fmt.Errorf("failed to list applications: %w", appsRes.Error())
	}

	apps := *appsRes.Payload()
	swept := 0
	errors := 0

	for _, app := range apps {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(app.Name, "tf-test") {
			continue
		}

		// Skip apps already marked for deletion
		if app.State == "TO_DELETE" {
			log.Printf("[INFO] Skipping application already marked for deletion: %s (%s)", app.Name, app.ID)
			continue
		}

		log.Printf("[INFO] Deleting application: %s (%s)", app.Name, app.ID)
		delRes := tmp.DeleteApp(ctx, cc, org, app.ID)
		if delRes.HasError() {
			log.Printf("[ERROR] Failed to delete application %s: %v", app.ID, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d applications (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping applications", errors)
	}
	return nil
}

// sweepAddons removes all test addons
func SweepAddons(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig())
	org := os.Getenv("ORGANISATION")

	if org == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping addons in organization: %s", org)

	// List all addons
	addonsRes := tmp.ListAddons(ctx, cc, org)
	if addonsRes.HasError() {
		return fmt.Errorf("failed to list addons: %w", addonsRes.Error())
	}

	addons := addonsRes.Payload()
	swept := 0
	errors := 0

	for _, addon := range *addons {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(addon.Name, "tf-test") {
			continue
		}

		log.Printf("[INFO] Deleting addon: %s (%s) - provider: %s", addon.Name, addon.RealID, addon.Provider.ID)
		delRes := tmp.DeleteAddon(ctx, cc, org, addon.RealID)
		if delRes.HasError() {
			// If addon is already being deleted or doesn't exist, skip the error
			if strings.Contains(delRes.Error().Error(), "not found") ||
				strings.Contains(delRes.Error().Error(), "TO_DELETE") {
				log.Printf("[INFO] Addon %s already deleted or being deleted", addon.RealID)
				continue
			}
			log.Printf("[ERROR] Failed to delete addon %s: %v", addon.RealID, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d addons (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping addons", errors)
	}
	return nil
}
