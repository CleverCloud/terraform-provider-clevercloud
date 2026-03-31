package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

// OAuthConsumerRequest represents the request payload for creating/updating an OAuth consumer
type OAuthConsumerRequest struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	BaseURL     string                     `json:"baseUrl"`
	LogoURL     string                     `json:"picture"`
	WebsiteURL  string                     `json:"url"`
	Rights      OAuthConsumerRightsRequest `json:"rights"`
}

// OAuthConsumerRightsRequest represents the rights for an OAuth consumer
type OAuthConsumerRightsRequest struct {
	AccessOrganisations                      bool `json:"access_organisations"`
	AccessOrganisationsBills                 bool `json:"access_organisations_bills"`
	AccessOrganisationsConsumptionStatistics bool `json:"access_organisations_consumption_statistics"`
	AccessOrganisationsCreditCount           bool `json:"access_organisations_credit_count"`
	AccessPersonalInformation                bool `json:"access_personal_information"`
	ManageOrganisations                      bool `json:"manage_organisations"`
	ManageOrganisationsApplications          bool `json:"manage_organisations_applications"`
	ManageOrganisationsMembers               bool `json:"manage_organisations_members"`
	ManageOrganisationsServices              bool `json:"manage_organisations_services"`
	ManagePersonalInformation                bool `json:"manage_personal_information"`
	ManageSSHKeys                            bool `json:"manage_ssh_keys"`
}

// OAuthConsumerResponse represents the response from the API
type OAuthConsumerResponse struct {
	Key         string                      `json:"key"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	BaseURL     string                      `json:"baseUrl"`
	LogoURL     string                      `json:"picture"`
	WebsiteURL  string                      `json:"url"`
	Rights      OAuthConsumerRightsResponse `json:"rights"`
}

// OAuthConsumerRightsResponse represents the rights from the API response
type OAuthConsumerRightsResponse struct {
	Almighty                                 bool `json:"almighty"`
	AccessOrganisations                      bool `json:"access_organisations"`
	AccessOrganisationsBills                 bool `json:"access_organisations_bills"`
	AccessOrganisationsConsumptionStatistics bool `json:"access_organisations_consumption_statistics"`
	AccessOrganisationsCreditCount           bool `json:"access_organisations_credit_count"`
	AccessPersonalInformation                bool `json:"access_personal_information"`
	ManageOrganisations                      bool `json:"manage_organisations"`
	ManageOrganisationsApplications          bool `json:"manage_organisations_applications"`
	ManageOrganisationsMembers               bool `json:"manage_organisations_members"`
	ManageOrganisationsServices              bool `json:"manage_organisations_services"`
	ManagePersonalInformation                bool `json:"manage_personal_information"`
	ManageSSHKeys                            bool `json:"manage_ssh_keys"`
}

// OAuthConsumerSecretResponse represents the secret response
type OAuthConsumerSecretResponse struct {
	Secret string `json:"secret"`
}

// CreateOAuthConsumer creates a new OAuth consumer
func CreateOAuthConsumer(ctx context.Context, c *client.Client, orgID string, req OAuthConsumerRequest) client.Response[OAuthConsumerResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/consumers", orgID)
	return client.Post[OAuthConsumerResponse](ctx, c, path, req)
}

// GetOAuthConsumer retrieves an OAuth consumer
func GetOAuthConsumer(ctx context.Context, c *client.Client, orgID string, consumerKey string) client.Response[OAuthConsumerResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/consumers/%s", orgID, consumerKey)
	return client.Get[OAuthConsumerResponse](ctx, c, path)
}

// GetOAuthConsumerSecret retrieves the secret for an OAuth consumer
func GetOAuthConsumerSecret(ctx context.Context, c *client.Client, orgID string, consumerKey string) client.Response[OAuthConsumerSecretResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/consumers/%s/secret", orgID, consumerKey)
	return client.Get[OAuthConsumerSecretResponse](ctx, c, path)
}

// UpdateOAuthConsumer updates an existing OAuth consumer
func UpdateOAuthConsumer(ctx context.Context, c *client.Client, orgID string, consumerKey string, req OAuthConsumerRequest) client.Response[OAuthConsumerResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/consumers/%s", orgID, consumerKey)
	return client.Put[OAuthConsumerResponse](ctx, c, path, req)
}

// DeleteOAuthConsumer deletes an OAuth consumer
// The API returns no content (204 or empty response)
func DeleteOAuthConsumer(ctx context.Context, c *client.Client, orgID string, consumerKey string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/consumers/%s", orgID, consumerKey)
	return client.Delete[client.Nothing](ctx, c, path)
}

// ListOAuthConsumers retrieves all OAuth consumers for an organization
func ListOAuthConsumers(ctx context.Context, c *client.Client, orgID string) client.Response[[]OAuthConsumerResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/consumers", orgID)
	return client.Get[[]OAuthConsumerResponse](ctx, c, path)
}
