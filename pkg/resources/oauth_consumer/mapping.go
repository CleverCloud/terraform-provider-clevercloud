package oauth_consumer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.clever-cloud.com/terraform-provider/pkg"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

// mapping.go holds every API-to-state mapper for clevercloud_oauth_consumer.
// See CONTRIBUTING.md § "API → state mapping".

// FromConsumer maps the consumer view, rights included.
//
// The secret is not here: it comes from a separate endpoint, hence FromSecret.
func (c *OAuthConsumer) FromConsumer(ctx context.Context, api *tmp.OAuthConsumerResponse, diags *diag.Diagnostics) *OAuthConsumer {
	if c == nil || api == nil {
		return c
	}

	c.Name = pkg.FromStr(api.Name)
	c.Description = pkg.FromStr(api.Description)
	c.BaseURL = pkg.FromStr(api.BaseURL)
	c.LogoURL = pkg.FromStr(api.LogoURL)
	c.WebsiteURL = pkg.FromStr(api.WebsiteURL)
	c.Rights = rightsResponseToSet(ctx, api.Rights, diags)

	return c
}

// FromSecret maps the separate secret endpoint.
func (c *OAuthConsumer) FromSecret(ctx context.Context, api *tmp.OAuthConsumerSecretResponse, diags *diag.Diagnostics) *OAuthConsumer {
	if c == nil || api == nil {
		return c
	}

	c.Secret = pkg.FromStr(api.Secret)

	return c
}
