package impl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.ProviderWithConfigValidators = (*Provider)(nil)

// ConfigValidators enforces the mutual exclusion between the API token (Bearer)
// authentication and the OAuth1 one: a configuration providing both is ambiguous.
func (p *Provider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		providervalidator.Conflicting(
			path.MatchRoot("api_token"),
			path.MatchRoot("token"),
		),
		providervalidator.Conflicting(
			path.MatchRoot("api_token"),
			path.MatchRoot("secret"),
		),
		providervalidator.Conflicting(
			path.MatchRoot("api_token"),
			path.MatchRoot("consumer_key"),
		),
		providervalidator.Conflicting(
			path.MatchRoot("api_token"),
			path.MatchRoot("consumer_secret"),
		),
	}
}
