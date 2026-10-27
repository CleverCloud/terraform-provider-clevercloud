package sweepers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

// sweepNetworkgroups removes all test networkgroups
func SweepNetworkgroups(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping networkgroups in organization: %s", tests.ORGANISATION)

	// List all networkgroups
	ngsRes := pkg.SDK(cc).V4().Networkgroups().Organisations().Ownerid(tests.ORGANISATION).Networkgroups().Listnetworkgroups(ctx)
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
		delRes := pkg.SDK(cc).V4().Networkgroups().Organisations().Ownerid(tests.ORGANISATION).Networkgroups().Networkgroupid(ng.ID).Deletenetworkgroup(ctx)
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
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping applications in organization: %s", tests.ORGANISATION)

	// List all applications
	appsRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Listorganisationapplications(ctx)
	if appsRes.HasError() {
		return fmt.Errorf("failed to list applications: %w", appsRes.Error())
	}

	apps := *appsRes.Payload()
	swept := 0
	errors := 0

	for _, app := range apps {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(app.Name, "tf-") {
			continue
		}

		// Skip apps already marked for deletion
		if app.State == "TO_DELETE" {
			log.Printf("[INFO] Skipping application already marked for deletion: %s (%s)", app.Name, app.ID)
			continue
		}

		log.Printf("[INFO] Deleting application: %s (%s)", app.Name, app.ID)
		delRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Applications().Appid(app.ID).Deleteorganisationapplication(ctx)
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
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping addons in organization: %s", tests.ORGANISATION)

	// List all addons
	addonsRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Addons().Getaddonsbyorgaid(ctx)
	if addonsRes.HasError() {
		return fmt.Errorf("failed to list addons: %w", addonsRes.Error())
	}

	addonList := addonsRes.Payload()
	swept := 0
	errors := 0

	for _, addon := range *addonList {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(pkg.Deref(addon.Name), "tf-test") {
			continue
		}

		log.Printf("[INFO] Deleting addon: %s (%s) - provider: %s", pkg.Deref(addon.Name), pkg.Deref(addon.RealID), addon.Provider.ID)
		delRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Addons().Addonid(pkg.Deref(addon.RealID)).Deprovisionaddonbyorgaandaddonid(ctx)
		if delRes.IsNotFoundError() {
			log.Printf("[INFO] Addon %s already deleted", pkg.Deref(addon.RealID))
			continue
		}
		if delRes.HasError() {
			// If addon is already being deleted or doesn't exist, skip the error
			if strings.Contains(delRes.Error().Error(), "TO_DELETE") {
				log.Printf("[INFO] Addon %s already being deleted", pkg.Deref(addon.RealID))
				continue
			}
			log.Printf("[ERROR] Failed to delete addon %s: %v", pkg.Deref(addon.RealID), delRes.Error())
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

// sweepKubernetes removes all test Kubernetes clusters
func SweepKubernetes(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping Kubernetes clusters in organization: %s", tests.ORGANISATION)

	// List all Kubernetes clusters
	clustersRes := pkg.SDK(cc).V4().Kubernetes().Organisations().Ownerid(tests.ORGANISATION).Clusters().Listkubernetesclusters(ctx)
	if clustersRes.HasError() {
		return fmt.Errorf("failed to list Kubernetes clusters: %w", clustersRes.Error())
	}

	clusters := *clustersRes.Payload()
	swept := 0
	errors := 0

	for _, cluster := range clusters {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(cluster.Name, "tf-test") {
			continue
		}

		// Skip clusters already marked for deletion
		if cluster.Status == "DELETED" || cluster.Status == "DELETING" {
			log.Printf("[INFO] Skipping Kubernetes cluster already marked for deletion: %s (%s) - Status: %s", cluster.Name, cluster.ID, cluster.Status)
			continue
		}

		log.Printf("[INFO] Deleting Kubernetes cluster: %s (%s) - Status: %s", cluster.Name, cluster.ID, cluster.Status)
		delRes := pkg.SDK(cc).V4().Kubernetes().Organisations().Ownerid(tests.ORGANISATION).Clusters().Kubernetesid(cluster.ID).Deletekubernetescluster(ctx)
		if delRes.IsNotFoundError() {
			log.Printf("[INFO] Kubernetes cluster %s already deleted", cluster.ID)
			continue
		}
		if delRes.HasError() {
			log.Printf("[ERROR] Failed to delete Kubernetes cluster %s: %v", cluster.ID, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d Kubernetes clusters (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping Kubernetes clusters", errors)
	}
	return nil
}

// sweepAddonProviders removes all test addon providers
func SweepAddonProviders(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping addon providers in organization: %s", tests.ORGANISATION)

	// List all addon providers
	providersRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Addonproviders().Getprovidersinfo(ctx)
	if providersRes.HasError() {
		return fmt.Errorf("failed to list addon providers: %w", providersRes.Error())
	}

	providers := *providersRes.Payload()
	swept := 0
	errors := 0

	for _, provider := range providers {
		// Only delete test resources (those starting with tf-test- or tftest)
		if !strings.HasPrefix(provider.ID, "tf-test-") && !strings.HasPrefix(provider.ID, "tftest") {
			continue
		}

		log.Printf("[INFO] Deleting addon provider: %s (%s)", pkg.Deref(provider.Name), provider.ID)
		delRes := tmp.DeleteAddonProvider(ctx, cc, tests.ORGANISATION, provider.ID)
		if delRes.IsNotFoundError() {
			log.Printf("[INFO] Addon provider %s already deleted", provider.ID)
			continue
		}
		if delRes.HasError() {
			log.Printf("[ERROR] Failed to delete addon provider %s: %v", provider.ID, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d addon providers (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping addon providers", errors)
	}
	return nil
}

type esClusterSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SweepElasticsearchClusters removes all test Elasticsearch clusters
func SweepElasticsearchClusters(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping Elasticsearch clusters in organization: %s", tests.ORGANISATION)

	path := fmt.Sprintf("/v4/elasticsearch/organisations/%s/clusters", tests.ORGANISATION)
	listRes := client.Get[[]esClusterSummary](ctx, cc, path)
	if listRes.HasError() {
		return fmt.Errorf("failed to list Elasticsearch clusters: %w", listRes.Error())
	}

	clusters := *listRes.Payload()
	swept := 0
	errors := 0

	for _, cluster := range clusters {
		if !strings.HasPrefix(cluster.Name, "tf-test") {
			continue
		}

		log.Printf("[INFO] Deleting Elasticsearch cluster: %s (%s)", cluster.Name, cluster.ID)
		delRes := client.Delete[client.Nothing](ctx, cc, fmt.Sprintf("%s/%s", path, cluster.ID))
		if delRes.IsNotFoundError() {
			log.Printf("[INFO] Elasticsearch cluster %s already deleted", cluster.ID)
			continue
		}
		if delRes.HasError() {
			log.Printf("[ERROR] Failed to delete Elasticsearch cluster %s: %v", cluster.ID, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d Elasticsearch clusters (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping Elasticsearch clusters", errors)
	}
	return nil
}

// sweepOAuthConsumers removes all test OAuth consumers
func SweepOAuthConsumers(region string) error {
	ctx := context.Background()
	cc := client.New(client.WithAutoOauthConfig(), client.WithRetryPolicy(pkg.RetryServerErrors))

	if tests.ORGANISATION == "" {
		return fmt.Errorf("ORGANISATION environment variable not set")
	}

	log.Printf("[INFO] Sweeping OAuth consumers in organization: %s", tests.ORGANISATION)

	// List all OAuth consumers
	consumersRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Consumers().Listorganisationconsumers(ctx)
	if consumersRes.HasError() {
		return fmt.Errorf("failed to list OAuth consumers: %w", consumersRes.Error())
	}

	consumers := consumersRes.Payload()
	swept := 0
	errors := 0

	for _, consumer := range *consumers {
		// Only delete test resources (those starting with tf-test)
		if !strings.HasPrefix(pkg.Deref(consumer.Name), "tf-test") {
			continue
		}

		log.Printf("[INFO] Deleting OAuth consumer: %s (key: %s)", pkg.Deref(consumer.Name), consumer.Key)
		delRes := pkg.SDK(cc).V2().Organisations().ID(tests.ORGANISATION).Consumers().Key(consumer.Key).Deleteorganisationconsumer(ctx)
		if delRes.IsNotFoundError() {
			log.Printf("[INFO] OAuth consumer %s already deleted", consumer.Key)
			continue
		}
		if delRes.HasError() {
			log.Printf("[ERROR] Failed to delete OAuth consumer %s: %v", consumer.Key, delRes.Error())
			errors++
			continue
		}
		swept++
	}

	log.Printf("[INFO] Swept %d OAuth consumers (errors: %d)", swept, errors)
	if errors > 0 {
		return fmt.Errorf("encountered %d errors while sweeping OAuth consumers", errors)
	}
	return nil
}
