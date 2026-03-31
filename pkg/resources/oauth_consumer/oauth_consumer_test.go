package oauth_consumer_test

import (
	"context"
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"go.clever-cloud.com/terraform-provider/pkg/helper"
	"go.clever-cloud.com/terraform-provider/pkg/tests"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
	"go.clever-cloud.dev/client"
)

func TestAccOAuthConsumer_basic(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-oauth")
	rNameEdited := rName + "-edited"
	fullName := fmt.Sprintf("clevercloud_oauth_consumer.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	oauthConsumerBlock := helper.NewRessource(
		"clevercloud_oauth_consumer",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":        rName,
			"description": "Test OAuth consumer",
			"base_url":    "https://api.example.com",
			"logo_url":    "https://example.com/logo.png",
			"website_url": "https://example.com",
			"rights": []string{
				"access_organisations",
				"manage_organisations_applications",
			},
		}))

	retrieveOAuthConsumer := func(ctx context.Context, consumerKey string) (*tmp.OAuthConsumerResponse, error) {
		res := tmp.GetOAuthConsumer(ctx, cc, tests.ORGANISATION, consumerKey)
		if res.IsNotFoundError() {
			return nil, fmt.Errorf("Unable to find OAuth consumer by key: %s", consumerKey)
		}
		if res.HasError() {
			return nil, fmt.Errorf("Unexpected error: %s", res.Error().Error())
		}
		return res.Payload(), nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(oauthConsumerBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("secret")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("description"), knownvalue.StringExact("Test OAuth consumer")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("base_url"), knownvalue.StringExact("https://api.example.com")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("logo_url"), knownvalue.StringExact("https://example.com/logo.png")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("website_url"), knownvalue.StringExact("https://example.com")),
				tests.NewCheckRemoteResource(fullName, retrieveOAuthConsumer, func(ctx context.Context, consumerKey string, state *tfjson.State, consumer *tmp.OAuthConsumerResponse) error {
					if consumer.Name != rName {
						return tests.AssertError("Bad consumer name", consumer.Name, rName)
					}
					if consumer.Description != "Test OAuth consumer" {
						return tests.AssertError("Bad consumer description", consumer.Description, "Test OAuth consumer")
					}
					if consumer.BaseURL != "https://api.example.com" {
						return tests.AssertError("Bad consumer base_url", consumer.BaseURL, "https://api.example.com")
					}
					if consumer.LogoURL != "https://example.com/logo.png" {
						return tests.AssertError("Bad consumer logo_url", consumer.LogoURL, "https://example.com/logo.png")
					}
					if consumer.WebsiteURL != "https://example.com" {
						return tests.AssertError("Bad consumer website_url", consumer.WebsiteURL, "https://example.com")
					}
					if !consumer.Rights.AccessOrganisations {
						return tests.AssertError("Bad consumer rights: access_organisations should be true", consumer.Rights.AccessOrganisations, true)
					}
					if !consumer.Rights.ManageOrganisationsApplications {
						return tests.AssertError("Bad consumer rights: manage_organisations_applications should be true", consumer.Rights.ManageOrganisationsApplications, true)
					}
					return nil
				}),
			},
		}, {
			ResourceName: rName,
			Config:       providerBlock.Append(oauthConsumerBlock.SetOneValue("name", rNameEdited)).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rNameEdited)),
			},
		}},
	})
}

func TestAccOAuthConsumer_withAllRights(t *testing.T) {
	ctx := t.Context()
	t.Parallel()
	cc := client.New(client.WithAutoOauthConfig())
	rName := acctest.RandomWithPrefix("tf-test-oauth-all")
	fullName := fmt.Sprintf("clevercloud_oauth_consumer.%s", rName)
	providerBlock := helper.NewProvider("clevercloud").SetOrganisation(tests.ORGANISATION)
	oauthConsumerBlock := helper.NewRessource(
		"clevercloud_oauth_consumer",
		rName,
		helper.SetKeyValues(map[string]any{
			"name":        rName,
			"description": "Test OAuth consumer with all rights",
			"base_url":    "https://api.example.com",
			"website_url": "https://example.com",
			"logo_url":    "https://example.com/logo.png",
			"rights": []string{
				"access_organisations",
				"access_organisations_bills",
				"access_organisations_consumption_statistics",
				"access_organisations_credit_count",
				"access_personal_information",
				"manage_organisations",
				"manage_organisations_applications",
				"manage_organisations_members",
				"manage_organisations_services",
				"manage_personal_information",
				"manage_ssh_keys",
			},
		}))

	retrieveOAuthConsumer := func(ctx context.Context, consumerKey string) (*tmp.OAuthConsumerResponse, error) {
		res := tmp.GetOAuthConsumer(ctx, cc, tests.ORGANISATION, consumerKey)
		if res.IsNotFoundError() {
			return nil, fmt.Errorf("Unable to find OAuth consumer by key: %s", consumerKey)
		}
		if res.HasError() {
			return nil, fmt.Errorf("Unexpected error: %s", res.Error().Error())
		}
		return res.Payload(), nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6Provider,
		PreCheck:                 tests.ExpectOrganisation(t),
		CheckDestroy:             tests.CheckDestroy(ctx),
		Steps: []resource.TestStep{{
			ResourceName: rName,
			Config:       providerBlock.Append(oauthConsumerBlock).String(),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("id"), knownvalue.NotNull()),
				statecheck.ExpectSensitiveValue(fullName, tfjsonpath.New("secret")),
				statecheck.ExpectKnownValue(fullName, tfjsonpath.New("name"), knownvalue.StringExact(rName)),
				tests.NewCheckRemoteResource(fullName, retrieveOAuthConsumer, func(ctx context.Context, consumerKey string, state *tfjson.State, consumer *tmp.OAuthConsumerResponse) error {
					if consumer.Name != rName {
						return tests.AssertError("Bad consumer name", consumer.Name, rName)
					}
					if consumer.Description != "Test OAuth consumer with all rights" {
						return tests.AssertError("Bad consumer description", consumer.Description, "Test OAuth consumer with all rights")
					}
					// Verify all rights are set to true
					if !consumer.Rights.AccessOrganisations {
						return tests.AssertError("Bad consumer rights: access_organisations should be true", consumer.Rights.AccessOrganisations, true)
					}
					if !consumer.Rights.AccessOrganisationsBills {
						return tests.AssertError("Bad consumer rights: access_organisations_bills should be true", consumer.Rights.AccessOrganisationsBills, true)
					}
					if !consumer.Rights.AccessOrganisationsConsumptionStatistics {
						return tests.AssertError("Bad consumer rights: access_organisations_consumption_statistics should be true", consumer.Rights.AccessOrganisationsConsumptionStatistics, true)
					}
					if !consumer.Rights.AccessOrganisationsCreditCount {
						return tests.AssertError("Bad consumer rights: access_organisations_credit_count should be true", consumer.Rights.AccessOrganisationsCreditCount, true)
					}
					if !consumer.Rights.AccessPersonalInformation {
						return tests.AssertError("Bad consumer rights: access_personal_information should be true", consumer.Rights.AccessPersonalInformation, true)
					}
					if !consumer.Rights.ManageOrganisations {
						return tests.AssertError("Bad consumer rights: manage_organisations should be true", consumer.Rights.ManageOrganisations, true)
					}
					if !consumer.Rights.ManageOrganisationsApplications {
						return tests.AssertError("Bad consumer rights: manage_organisations_applications should be true", consumer.Rights.ManageOrganisationsApplications, true)
					}
					if !consumer.Rights.ManageOrganisationsMembers {
						return tests.AssertError("Bad consumer rights: manage_organisations_members should be true", consumer.Rights.ManageOrganisationsMembers, true)
					}
					if !consumer.Rights.ManageOrganisationsServices {
						return tests.AssertError("Bad consumer rights: manage_organisations_services should be true", consumer.Rights.ManageOrganisationsServices, true)
					}
					if !consumer.Rights.ManagePersonalInformation {
						return tests.AssertError("Bad consumer rights: manage_personal_information should be true", consumer.Rights.ManagePersonalInformation, true)
					}
					if !consumer.Rights.ManageSSHKeys {
						return tests.AssertError("Bad consumer rights: manage_ssh_keys should be true", consumer.Rights.ManageSSHKeys, true)
					}
					return nil
				}),
			},
		}},
	})
}
