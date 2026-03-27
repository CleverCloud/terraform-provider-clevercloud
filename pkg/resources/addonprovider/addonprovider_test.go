package addonprovider_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

func TestAccAddonProvider_basic(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ap")
	fullName := fmt.Sprintf("clevercloud_addon_provider.%s", rName)

	// Generate secure random strings for password and sso_salt (min 35 chars)
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Config vars must be prefixed with the provider_id (uppercased, dashes replaced by underscores)
	// For provider_id "tf-test-ap-xxx", config vars must start with "TF_TEST_AP_XXX_"
	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_DATABASE_URL", providerIDUpper),
		fmt.Sprintf("%s_API_KEY", providerIDUpper),
	}

	addonProviderBlock := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider",
			"config_vars":         configVars,
			"password":            password,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("provider_id"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact("TF Test Addon Provider")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("config_vars"), knownvalue.SetExact([]knownvalue.Check{
					knownvalue.StringExact(configVars[0]),
					knownvalue.StringExact(configVars[1]),
				})),
				// regions is computed from API and will contain all supported regions
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("regions"), knownvalue.NotNull()),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("production_base_url"), knownvalue.StringExact("https://example.com/clevercloud/resources")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("production_sso_url"), knownvalue.StringExact("https://example.com/clevercloud/sso/login")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("test_base_url"), knownvalue.StringExact("https://test.example.com/clevercloud/resources")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("test_sso_url"), knownvalue.StringExact("https://test.example.com/clevercloud/sso/login")),
			},
			Check: func(state *terraform.State) error {
				providerRes := tmp.GetAddonProvider(ctx, cc, tests.ORGANISATION, rName)
				if providerRes.HasError() {
					return fmt.Errorf("failed to get addon provider from API: %w", providerRes.Error())
				}
				provider := *providerRes.Payload()

				// Verify basic fields that the API returns
				// Note: The API does not return sensitive fields (password, sso_salt)
				// or configuration details (config_vars, URLs) in the GET response
				if provider.ID != rName {
					return fmt.Errorf("expected provider ID %q, got %q", rName, provider.ID)
				}
				if provider.Name != "TF Test Addon Provider" {
					return fmt.Errorf("expected provider name %q, got %q", "TF Test Addon Provider", provider.Name)
				}

				// Verify regions - API returns all supported regions, not just the one we specified
				if len(provider.Regions) == 0 {
					return fmt.Errorf("expected at least one region, got 0")
				}

				// Verify the status field exists
				if provider.Status == "" {
					return fmt.Errorf("expected 'status' to be a non-empty string, got empty string")
				}

				return nil
			},
		}},
	})
}

func TestAccAddonProvider_passwordValidation(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-ap")

	// Short password (less than 35 chars)
	shortPassword := "tooshort"
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_VAR", providerIDUpper),
	}

	addonProviderBlock := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider",
			"config_vars":         configVars,
			"password":            shortPassword,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ExpectError:  regexp.MustCompile("string length must be at least 35"),
		}},
	})
}

func TestAccAddonProvider_ssoSaltValidation(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-ap")
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	// Short sso_salt (less than 35 chars)
	shortSsoSalt := "tooshort"

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_VAR", providerIDUpper),
	}

	addonProviderBlock := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider",
			"config_vars":         configVars,
			"password":            password,
			"sso_salt":            shortSsoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ExpectError:  regexp.MustCompile("string length must be at least 35"),
		}},
	})
}

func TestAccAddonProvider_httpsValidation(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test-ap")
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_VAR", providerIDUpper),
	}

	testCases := []struct {
		name        string
		urlField    string
		urlValue    string
		expectError string
	}{
		{
			name:        "http_production_base_url",
			urlField:    "production_base_url",
			urlValue:    "http://example.com/clevercloud/resources",
			expectError: "Invalid URL Scheme",
		},
		{
			name:        "http_production_sso_url",
			urlField:    "production_sso_url",
			urlValue:    "http://example.com/clevercloud/sso",
			expectError: "Invalid URL Scheme",
		},
		{
			name:        "http_test_base_url",
			urlField:    "test_base_url",
			urlValue:    "http://example.com/clevercloud/resources",
			expectError: "Invalid URL Scheme",
		},
		{
			name:        "http_test_sso_url",
			urlField:    "test_sso_url",
			urlValue:    "http://example.com/clevercloud/sso",
			expectError: "Invalid URL Scheme",
		},
		{
			name:        "ftp_url",
			urlField:    "production_base_url",
			urlValue:    "ftp://example.com/resources",
			expectError: "Invalid URL Scheme",
		},
		{
			name:        "invalid_url",
			urlField:    "production_base_url",
			urlValue:    "not-a-valid-url",
			expectError: "Invalid URL Scheme",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			urlValues := map[string]any{
				"production_base_url": "https://example.com/clevercloud/resources",
				"production_sso_url":  "https://example.com/clevercloud/sso/login",
				"test_base_url":       "https://test.example.com/clevercloud/resources",
				"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
			}
			// Override the specific field being tested
			urlValues[tc.urlField] = tc.urlValue

			addonProviderBlock := helper.NewRessource(
				"clevercloud_addon_provider",
				rName,
				helper.SetKeyValues(map[string]any{
					"provider_id":         rName,
					"name":                "TF Test Addon Provider",
					"config_vars":         configVars,
					"password":            password,
					"sso_salt":            ssoSalt,
					"production_base_url": urlValues["production_base_url"],
					"production_sso_url":  urlValues["production_sso_url"],
					"test_base_url":       urlValues["test_base_url"],
					"test_sso_url":        urlValues["test_sso_url"],
				}),
			)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: tests.ProtoV6Provider,
				PreCheck:                 tests.ExpectOrganisation(t),
				Steps: []resource.TestStep{{
					ResourceName: rName,
					Config:       providerBlock.Append(addonProviderBlock).String(),
					ExpectError:  regexp.MustCompile(tc.expectError),
				}},
			})
		})
	}
}

func TestAccAddonProvider_configVarsValidation(t *testing.T) {
	ctx := t.Context()
	rName := acctest.RandomWithPrefix("tf-test-ap")
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Invalid config_vars - not prefixed with provider_id
	invalidConfigVars := []string{
		"WRONG_PREFIX_VAR",
	}

	addonProviderBlock := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider",
			"config_vars":         invalidConfigVars,
			"password":            password,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ExpectError:  regexp.MustCompile("Invalid config_vars prefix"),
		}},
	})
}

func TestAccAddonProvider_withFeatures(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ap")
	fullName := fmt.Sprintf("clevercloud_addon_provider.%s", rName)

	// Generate secure random strings for password and sso_salt (min 35 chars)
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Config vars must be prefixed with the provider_id (uppercased, dashes replaced by underscores)
	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_DATABASE_URL", providerIDUpper),
		fmt.Sprintf("%s_API_KEY", providerIDUpper),
	}

	addonProviderBlock := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider with Features",
			"config_vars":         configVars,
			"password":            password,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
		helper.SetNestedBlockValues("feature", []map[string]string{
			{"name": "disk_size", "type": "FILESIZE"},
			{"name": "connection_limit", "type": "NUMBER"},
			{"name": "awesome_features", "type": "BOOLEAN"},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("provider_id"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact("TF Test Addon Provider with Features")),
				// Verify features exist (3 features)
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("feature"), knownvalue.SetSizeExact(3)),
			},
			Check: func(state *terraform.State) error {
				featuresRes := tmp.ListAddonProviderFeatures(ctx, cc, tests.ORGANISATION, rName)
				if featuresRes.HasError() {
					return fmt.Errorf("failed to get providerfeatures : %w", featuresRes.Error())
				}
				features := *featuresRes.Payload()

				if len(features) != 3 {
					return fmt.Errorf("Expect 3 features, got %d", len(features))
				}

				return nil
			},
		}},
	})
}

func TestAccAddonProvider_withPlans(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ap")
	fullName := fmt.Sprintf("clevercloud_addon_provider.%s", rName)

	// Generate secure random strings for password and sso_salt (min 35 chars)
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	// Config vars must be prefixed with the provider_id (uppercased, dashes replaced by underscores)
	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_DATABASE_URL", providerIDUpper),
		fmt.Sprintf("%s_API_KEY", providerIDUpper),
	}

	addonProviderBlock := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider with Plans",
			"config_vars":         configVars,
			"password":            password,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
		helper.SetNestedBlockValues("plan", []map[string]string{
			{"name": "Free Plan", "slug": "free", "price": "0"},
			{"name": "Basic Plan", "slug": "basic", "price": "9.99"},
			{"name": "Premium Plan", "slug": "premium", "price": "29.99"},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("provider_id"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact("TF Test Addon Provider with Plans")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.SetSizeExact(3)),
			},
			Check: func(state *terraform.State) error {
				plansRes := tmp.ListAddonProviderPlans(ctx, cc, tests.ORGANISATION, rName)
				if plansRes.HasError() {
					return fmt.Errorf("failed to get provider plans: %w", plansRes.Error())
				}
				plans := *plansRes.Payload()

				if len(plans) != 3 {
					return fmt.Errorf("expect 3 plans, got %d", len(plans))
				}

				// Verify plan details
				plansBySlug := make(map[string]tmp.AddonProviderPlanView)
				for _, plan := range plans {
					plansBySlug[plan.Slug] = plan
				}

				// Check Free Plan
				if freePlan, ok := plansBySlug["free"]; !ok {
					return fmt.Errorf("free plan not found")
				} else if freePlan.Name != "Free Plan" || freePlan.Price != 0 {
					return fmt.Errorf("free plan has incorrect values: name=%s, price=%f", freePlan.Name, freePlan.Price)
				}

				// Check Basic Plan
				if basicPlan, ok := plansBySlug["basic"]; !ok {
					return fmt.Errorf("basic plan not found")
				} else if basicPlan.Name != "Basic Plan" || basicPlan.Price != 9.99 {
					return fmt.Errorf("basic plan has incorrect values: name=%s, price=%f", basicPlan.Name, basicPlan.Price)
				}

				// Check Premium Plan
				if premiumPlan, ok := plansBySlug["premium"]; !ok {
					return fmt.Errorf("premium plan not found")
				} else if premiumPlan.Name != "Premium Plan" || premiumPlan.Price != 29.99 {
					return fmt.Errorf("premium plan has incorrect values: name=%s, price=%f", premiumPlan.Name, premiumPlan.Price)
				}

				return nil
			},
		}},
	})
}

func TestAccAddonProvider_plansWithFeatures(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ap")
	fullName := fmt.Sprintf("clevercloud_addon_provider.%s", rName)

	// Generate secure random strings for password and sso_salt (min 35 chars)
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")

	// Config vars must be prefixed with the provider_id (uppercased, dashes replaced by underscores)
	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_DATABASE_URL", providerIDUpper),
		fmt.Sprintf("%s_API_KEY", providerIDUpper),
	}

	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	addonProviderBlock := helper.
		NewRessource(
			"clevercloud_addon_provider",
			rName,
			helper.SetKeyValues(map[string]any{
				"provider_id":         rName,
				"name":                "TF Test Addon Provider with Plans and Features",
				"config_vars":         configVars,
				"password":            password,
				"sso_salt":            ssoSalt,
				"production_base_url": "https://example.com/clevercloud/resources",
				"production_sso_url":  "https://example.com/clevercloud/sso/login",
				"test_base_url":       "https://test.example.com/clevercloud/resources",
				"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
			}),
			helper.SetNestedBlockValues("feature", []map[string]string{
				{"name": "disk_size", "type": "FILESIZE"},
				{"name": "connection_limit", "type": "NUMBER"},
			}),
		).
		AddNestedBlocks("plan", []helper.Block{
			helper.NewBlock(
				map[string]any{
					"name":  "Free Plan",
					"slug":  "free",
					"price": 0,
				},
				map[string]any{
					"features": []map[string]string{
						{"name": "disk_size", "value": "100"},
						{"name": "connection_limit", "value": "5"},
					},
				},
			),
			helper.NewBlock(
				map[string]any{
					"name":  "Premium Plan",
					"slug":  "premium",
					"price": 29.99,
				},
				map[string]any{
					"features": []map[string]string{
						{"name": "disk_size", "value": "1000"},
						{"name": "connection_limit", "value": "100"},
					},
				},
			),
		})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(addonProviderBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("provider_id"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact("TF Test Addon Provider with Plans and Features")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("feature"), knownvalue.SetSizeExact(2)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.SetSizeExact(2)),
			},
			Check: func(state *terraform.State) error {
				plansRes := tmp.ListAddonProviderPlans(ctx, cc, tests.ORGANISATION, rName)
				if plansRes.HasError() {
					return fmt.Errorf("failed to get provider plans: %w", plansRes.Error())
				}
				plans := *plansRes.Payload()

				if len(plans) != 2 {
					return fmt.Errorf("expect 2 plans, got %d", len(plans))
				}

				// Verify plan details and their feature values
				plansBySlug := make(map[string]tmp.AddonProviderPlanView)
				for _, plan := range plans {
					plansBySlug[plan.Slug] = plan
				}

				// Check Free Plan
				freePlan, ok := plansBySlug["free"]
				if !ok {
					return fmt.Errorf("free plan not found")
				}
				if freePlan.Name != "Free Plan" || freePlan.Price != 0 {
					return fmt.Errorf("free plan has incorrect values: name=%s, price=%f", freePlan.Name, freePlan.Price)
				}

				// Verify free plan features
				freeFeaturesByName := make(map[string]tmp.AddonPlanFeatureView)
				for _, feature := range freePlan.Features {
					freeFeaturesByName[feature.Name] = feature
				}
				if diskSize, ok := freeFeaturesByName["disk_size"]; !ok || diskSize.Value != "100" {
					return fmt.Errorf("free plan disk_size incorrect: expected '100', got '%s'", diskSize.Value)
				}
				if connLimit, ok := freeFeaturesByName["connection_limit"]; !ok || connLimit.Value != "5" {
					return fmt.Errorf("free plan connection_limit incorrect: expected '5', got '%s'", connLimit.Value)
				}

				// Check Premium Plan
				premiumPlan, ok := plansBySlug["premium"]
				if !ok {
					return fmt.Errorf("premium plan not found")
				}
				if premiumPlan.Name != "Premium Plan" || premiumPlan.Price != 29.99 {
					return fmt.Errorf("premium plan has incorrect values: name=%s, price=%f", premiumPlan.Name, premiumPlan.Price)
				}

				// Verify premium plan features
				premiumFeaturesByName := make(map[string]tmp.AddonPlanFeatureView)
				for _, feature := range premiumPlan.Features {
					premiumFeaturesByName[feature.Name] = feature
				}
				if diskSize, ok := premiumFeaturesByName["disk_size"]; !ok || diskSize.Value != "1000" {
					return fmt.Errorf("premium plan disk_size incorrect: expected '1000', got '%s'", diskSize.Value)
				}
				if connLimit, ok := premiumFeaturesByName["connection_limit"]; !ok || connLimit.Value != "100" {
					return fmt.Errorf("premium plan connection_limit incorrect: expected '100', got '%s'", connLimit.Value)
				}

				return nil
			},
		}},
	})
}

// TestAccAddonProvider_updateFeatures tests adding and removing features
func TestAccAddonProvider_updateFeatures(t *testing.T) {
	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ap")
	fullName := fmt.Sprintf("clevercloud_addon_provider.%s", rName)
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_DATABASE_URL", providerIDUpper),
		fmt.Sprintf("%s_API_KEY", providerIDUpper),
	}

	// Initial configuration with 2 features
	initialConfig := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider Update Features",
			"config_vars":         configVars,
			"password":            password,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
		helper.SetNestedBlockValues("feature", []map[string]string{
			{"name": "disk_size", "type": "FILESIZE"},
			{"name": "connection_limit", "type": "NUMBER"},
		}),
	)

	// Updated configuration: remove connection_limit, add backup_enabled
	updatedConfig := helper.NewRessource(
		"clevercloud_addon_provider",
		rName,
		helper.SetKeyValues(map[string]any{
			"provider_id":         rName,
			"name":                "TF Test Addon Provider Update Features",
			"config_vars":         configVars,
			"password":            password,
			"sso_salt":            ssoSalt,
			"production_base_url": "https://example.com/clevercloud/resources",
			"production_sso_url":  "https://example.com/clevercloud/sso/login",
			"test_base_url":       "https://test.example.com/clevercloud/resources",
			"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
		}),
		helper.SetNestedBlockValues("feature", []map[string]string{
			{"name": "disk_size", "type": "FILESIZE"},
			{"name": "backup_enabled", "type": "BOOLEAN"},
		}),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			Config: providerBlock.Append(initialConfig).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("feature"), knownvalue.SetSizeExact(2)),
			},
			Check: func(state *terraform.State) error {
				featuresRes := tmp.ListAddonProviderFeatures(ctx, cc, tests.ORGANISATION, rName)
				if featuresRes.HasError() {
					return fmt.Errorf("failed to get features: %w", featuresRes.Error())
				}
				features := *featuresRes.Payload()
				if len(features) != 2 {
					return fmt.Errorf("expected 2 features, got %d", len(features))
				}

				featureNames := make(map[string]bool)
				for _, f := range features {
					featureNames[f.Name] = true
				}

				if !featureNames["disk_size"] || !featureNames["connection_limit"] {
					return fmt.Errorf("expected disk_size and connection_limit features")
				}

				return nil
			},
		}, {
			Config: providerBlock.Append(updatedConfig).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("feature"), knownvalue.SetSizeExact(2)),
			},
			Check: func(state *terraform.State) error {
				featuresRes := tmp.ListAddonProviderFeatures(ctx, cc, tests.ORGANISATION, rName)
				if featuresRes.HasError() {
					return fmt.Errorf("failed to get features: %w", featuresRes.Error())
				}
				features := *featuresRes.Payload()
				if len(features) != 2 {
					return fmt.Errorf("expected 2 features, got %d", len(features))
				}

				featureNames := make(map[string]bool)
				for _, f := range features {
					featureNames[f.Name] = true
				}

				// Verify connection_limit was removed
				if featureNames["connection_limit"] {
					return fmt.Errorf("connection_limit should have been removed")
				}

				// Verify backup_enabled was added
				if !featureNames["backup_enabled"] {
					return fmt.Errorf("backup_enabled should have been added")
				}

				// Verify disk_size is still there
				if !featureNames["disk_size"] {
					return fmt.Errorf("disk_size should still exist")
				}

				return nil
			},
		},
		},
	})
}

// TestAccAddonProvider_updatePlans tests adding, updating, and removing plans
func TestAccAddonProvider_updatePlans(t *testing.T) {
	t.Skip("Skipping due to API bug: DELETE /plans/{id} returns 500 - See issue cc-api#847")

	ctx := t.Context()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-ap")
	fullName := fmt.Sprintf("clevercloud_addon_provider.%s", rName)
	password := acctest.RandomWithPrefix("pass") + acctest.RandomWithPrefix("")
	ssoSalt := acctest.RandomWithPrefix("salt") + acctest.RandomWithPrefix("")
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)

	providerIDUpper := strings.ToUpper(strings.ReplaceAll(rName, "-", "_"))
	configVars := []string{
		fmt.Sprintf("%s_DATABASE_URL", providerIDUpper),
		fmt.Sprintf("%s_API_KEY", providerIDUpper),
	}

	// Initial configuration with 2 plans and features
	initialConfig := helper.
		NewRessource(
			"clevercloud_addon_provider",
			rName,
			helper.SetKeyValues(map[string]any{
				"provider_id":         rName,
				"name":                "TF Test Addon Provider Update Plans",
				"config_vars":         configVars,
				"password":            password,
				"sso_salt":            ssoSalt,
				"production_base_url": "https://example.com/clevercloud/resources",
				"production_sso_url":  "https://example.com/clevercloud/sso/login",
				"test_base_url":       "https://test.example.com/clevercloud/resources",
				"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
			}),
			helper.SetNestedBlockValues("feature", []map[string]string{
				{"name": "disk_size", "type": "FILESIZE"},
				{"name": "connection_limit", "type": "NUMBER"},
			}),
		).
		AddNestedBlocks("plan", []helper.Block{
			helper.NewBlock(
				map[string]any{
					"name":  "Free Plan",
					"slug":  "free",
					"price": 0,
				},
				map[string]any{
					"features": []map[string]string{
						{"name": "disk_size", "value": "100"},
						{"name": "connection_limit", "value": "5"},
					},
				},
			),
			helper.NewBlock(
				map[string]any{
					"name":  "Starter Plan",
					"slug":  "starter",
					"price": 5.0,
				},
				map[string]any{
					"features": []map[string]string{
						{"name": "disk_size", "value": "500"},
						{"name": "connection_limit", "value": "50"},
					},
				},
			),
		})

	// Updated configuration: remove starter plan, update free plan values, add premium plan
	updatedConfig := helper.
		NewRessource(
			"clevercloud_addon_provider",
			rName,
			helper.SetKeyValues(map[string]any{
				"provider_id":         rName,
				"name":                "TF Test Addon Provider Update Plans",
				"config_vars":         configVars,
				"password":            password,
				"sso_salt":            ssoSalt,
				"production_base_url": "https://example.com/clevercloud/resources",
				"production_sso_url":  "https://example.com/clevercloud/sso/login",
				"test_base_url":       "https://test.example.com/clevercloud/resources",
				"test_sso_url":        "https://test.example.com/clevercloud/sso/login",
			}),
			helper.SetNestedBlockValues("feature", []map[string]string{
				{"name": "disk_size", "type": "FILESIZE"},
				{"name": "connection_limit", "type": "NUMBER"},
			}),
		).
		AddNestedBlocks("plan", []helper.Block{
			helper.NewBlock(
				map[string]any{
					"name":  "Free Plan",
					"slug":  "free",
					"price": 0,
				},
				map[string]any{
					"features": []map[string]string{
						{"name": "disk_size", "value": "200"},       // Updated from 100
						{"name": "connection_limit", "value": "10"}, // Updated from 5
					},
				},
			),
			helper.NewBlock(
				map[string]any{
					"name":  "Premium Plan",
					"slug":  "premium",
					"price": 20.0,
				},
				map[string]any{
					"features": []map[string]string{
						{"name": "disk_size", "value": "2000"},
						{"name": "connection_limit", "value": "200"},
					},
				},
			),
		})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			Config: providerBlock.Append(initialConfig).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.SetSizeExact(2)),
			},
			Check: func(state *terraform.State) error {
				plansRes := tmp.ListAddonProviderPlans(ctx, cc, tests.ORGANISATION, rName)
				if plansRes.HasError() {
					return fmt.Errorf("failed to get plans: %w", plansRes.Error())
				}
				plans := *plansRes.Payload()
				if len(plans) != 2 {
					return fmt.Errorf("expected 2 plans, got %d", len(plans))
				}

				plansBySlug := make(map[string]tmp.AddonProviderPlanView)
				for _, plan := range plans {
					plansBySlug[plan.Slug] = plan
				}

				// Verify free plan
				freePlan, ok := plansBySlug["free"]
				if !ok {
					return fmt.Errorf("free plan not found")
				}
				if freePlan.Name != "Free Plan" || freePlan.Price != 0 {
					return fmt.Errorf("free plan incorrect: name=%s, price=%f", freePlan.Name, freePlan.Price)
				}

				// Verify starter plan exists
				if _, ok := plansBySlug["starter"]; !ok {
					return fmt.Errorf("starter plan not found")
				}

				return nil
			},
		}, {
			Config: providerBlock.Append(updatedConfig).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("plan"), knownvalue.SetSizeExact(2)),
			},
			Check: func(state *terraform.State) error {
				plansRes := tmp.ListAddonProviderPlans(ctx, cc, tests.ORGANISATION, rName)
				if plansRes.HasError() {
					return fmt.Errorf("failed to get plans: %w", plansRes.Error())
				}
				plans := *plansRes.Payload()
				if len(plans) != 2 {
					return fmt.Errorf("expected 2 plans, got %d", len(plans))
				}

				plansBySlug := make(map[string]tmp.AddonProviderPlanView)
				for _, plan := range plans {
					plansBySlug[plan.Slug] = plan
				}

				// Verify starter was removed
				if _, ok := plansBySlug["starter"]; ok {
					return fmt.Errorf("starter plan should have been removed")
				}

				// Verify premium was added
				premiumPlan, ok := plansBySlug["premium"]
				if !ok {
					return fmt.Errorf("premium plan not found")
				}
				if premiumPlan.Name != "Premium Plan" || premiumPlan.Price != 20.0 {
					return fmt.Errorf("premium plan incorrect: name=%s, price=%f", premiumPlan.Name, premiumPlan.Price)
				}

				// Verify free plan was updated
				freePlan, ok := plansBySlug["free"]
				if !ok {
					return fmt.Errorf("free plan not found")
				}

				// Check updated feature values
				freeFeaturesByName := make(map[string]tmp.AddonPlanFeatureView)
				for _, feature := range freePlan.Features {
					freeFeaturesByName[feature.Name] = feature
				}

				if diskSize, ok := freeFeaturesByName["disk_size"]; !ok || diskSize.Value != "200" {
					return fmt.Errorf("free plan disk_size should be '200', got '%s'", diskSize.Value)
				}

				if connLimit, ok := freeFeaturesByName["connection_limit"]; !ok || connLimit.Value != "10" {
					return fmt.Errorf("free plan connection_limit should be '10', got '%s'", connLimit.Value)
				}

				return nil
			},
		},
		},
	})
}
