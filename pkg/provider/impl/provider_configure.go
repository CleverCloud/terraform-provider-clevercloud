package impl

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.com/terraform-provider/pkg/clevertools"
	"go.clever-cloud.dev/client"
)

type oauthCreds struct{ consumerKey, consumerSecret, token, secret string }

// resolveCreds mirrors go.clever-cloud.dev/client credential precedence (env, then
// clever-tools config) while also understanding the profiles format introduced in
// clever-tools 4.6.0, which the pinned client v0.1.7 cannot parse.
func resolveCreds() (*oauthCreds, *clevertools.Profile, diag.Diagnostics) {
	var diags diag.Diagnostics

	token, secret := os.Getenv("CC_OAUTH_TOKEN"), os.Getenv("CC_OAUTH_SECRET")
	if token != "" && secret != "" {
		return &oauthCreds{
			consumerKey:    firstNonEmpty(os.Getenv("CC_CONSUMER_KEY"), client.OAUTH_CONSUMER_KEY),
			consumerSecret: firstNonEmpty(os.Getenv("CC_CONSUMER_SECRET"), client.OAUTH_CONSUMER_SECRET),
			token:          token,
			secret:         secret,
		}, nil, diags
	}

	path := clevertools.ConfigFilePath()

	profile, err := clevertools.ActiveProfile(path)
	if err != nil {
		diags.AddError(
			"Invalid clever-tools configuration",
			fmt.Sprintf("Could not read credentials from %q: %s", path, err),
		)

		return nil, nil, diags
	}

	if profile == nil {
		return nil, nil, diags
	}

	consumerKey, consumerSecret := client.OAUTH_CONSUMER_KEY, client.OAUTH_CONSUMER_SECRET
	if o := profile.Overrides; o != nil {
		consumerKey = firstNonEmpty(o.OAuthConsumerKey, consumerKey)
		consumerSecret = firstNonEmpty(o.OAuthConsumerSecret, consumerSecret)
	}

	return &oauthCreds{
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		token:          profile.Token,
		secret:         profile.Secret,
	}, profile, diags
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

// isSetStr reports whether a config attribute holds a known, non-empty value.
func isSetStr(v types.String) bool {
	return !v.IsUnknown() && !v.IsNull() && v.ValueString() != ""
}

// resolveAPIToken returns the API token to use for Bearer authentication and
// whether it was read from the environment: the api_token provider parameter
// takes precedence over the CLEVER_API_TOKEN environment variable.
func resolveAPIToken(config ProviderData) (token string, fromEnv bool) {
	if isSetStr(config.APIToken) {
		return config.APIToken.ValueString(), false
	}

	return os.Getenv("CLEVER_API_TOKEN"), true
}

// bearerEndpoint returns the endpoint to use for API token (Bearer)
// authentication. API tokens are only accepted by the API bridge, so the
// endpoint defaults to it unless explicitly configured.
func bearerEndpoint(endpoint string) string {
	if endpoint == "" {
		return client.BRIDGE_API_ENDPOINT
	}

	return endpoint
}

// bearerClientOptions returns the client options enabling API token (Bearer)
// authentication.
func bearerClientOptions(endpoint, apiToken string) []func(*client.Client) {
	return []func(*client.Client){
		client.WithEndpoint(bearerEndpoint(endpoint)),
		client.WithBearerAuth(apiToken),
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ProviderData

	tflog.Debug(ctx, "configure provider...")

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.isNetwrkgroupsDisabled = config.DisableNetworkgroup.ValueBool()

	p.organization = os.Getenv("CC_ORGANISATION")
	if !config.Organisation.IsUnknown() && !config.Organisation.IsNull() {
		p.organization = config.Organisation.ValueString()
	}
	if p.organization == "" {
		resp.Diagnostics.AddError("Invalid provider configuration", "Organisation should be set by either the organisation parameter or by the CC_ORGANISATION environment variable")
		return
	}

	// Explicit provider parameters always win over environment variables: an
	// API token coming from CLEVER_API_TOKEN is ignored when OAuth1 tokens are
	// set in the provider block (the parameter-level conflict is already
	// rejected by ConfigValidators). Two competing environment variables are
	// ambiguous, so fail with an actionable message.
	apiToken, apiTokenFromEnv := resolveAPIToken(config)
	useBearer := apiToken != ""
	if useBearer && apiTokenFromEnv {
		if isSetStr(config.Token) || isSetStr(config.Secret) {
			useBearer = false
		} else if os.Getenv("CC_OAUTH_TOKEN") != "" && os.Getenv("CC_OAUTH_SECRET") != "" {
			resp.Diagnostics.AddError(
				"Conflicting CleverCloud credentials",
				"Both the CLEVER_API_TOKEN and the CC_OAUTH_TOKEN/CC_OAUTH_SECRET environment variables are set. Unset one of them, or select an authentication method explicitly with the api_token or token/secret provider parameters.",
			)
			return
		}
	}

	// Allow to get creds from CLI config directory or by injected variables
	var clientOptions []func(*client.Client)

	if useBearer {
		clientOptions = bearerClientOptions(config.Endpoint.ValueString(), apiToken)

		// API tokens cannot sign git operations: leave gitAuth unset, application
		// resources deploying code over git require OAuth1 credentials.
		tflog.Info(ctx, "using API token (Bearer) authentication, git deployments are unavailable")
	} else if !config.Endpoint.IsUnknown() && !config.Endpoint.IsNull() && config.Endpoint.ValueString() != "" {
		clientOptions = append(clientOptions, client.WithEndpoint(config.Endpoint.ValueString()))
	}

	// New branch: allow setting all OAuth1 params
	if !useBearer &&
		!config.ConsumerKey.IsUnknown() && !config.ConsumerKey.IsNull() && config.ConsumerKey.ValueString() != "" &&
		!config.ConsumerSecret.IsUnknown() && !config.ConsumerSecret.IsNull() && config.ConsumerSecret.ValueString() != "" &&
		!config.Token.IsUnknown() && !config.Token.IsNull() && config.Token.ValueString() != "" &&
		!config.Secret.IsUnknown() && !config.Secret.IsNull() && config.Secret.ValueString() != "" {
		clientOptions = append(clientOptions, client.WithOauthConfig(
			config.ConsumerKey.ValueString(),
			config.ConsumerSecret.ValueString(),
			config.Token.ValueString(),
			config.Secret.ValueString(),
		))
		p.gitAuth = &http.BasicAuth{Username: config.Token.ValueString(), Password: config.Secret.ValueString()}

	} else if !useBearer &&
		(config.Secret.IsUnknown() ||
			config.Token.IsUnknown() ||
			config.Secret.IsNull() ||
			config.Token.IsNull()) {
		creds, profile, diags := resolveCreds()
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if creds == nil {
			searched := clevertools.ConfigFilePath()
			if searched == "" {
				searched = "no clever-tools configuration file found"
			}

			resp.Diagnostics.AddError(
				"CleverCloud authentication empty",
				fmt.Sprintf(
					"No credentials found (%s).\n\nEither set the CC_OAUTH_TOKEN and CC_OAUTH_SECRET environment variables, run 'clever login', set the token and secret provider parameters, or provide an API token via the api_token parameter or the CLEVER_API_TOKEN environment variable.",
					searched,
				),
			)
			return
		}

		// An expired token only surfaces as an opaque 401 later on, and the legacy
		// config format may not carry an expiration date at all: warn, never fail.
		if profile.Expired() {
			resp.Diagnostics.AddWarning(
				"CleverCloud credentials expired",
				fmt.Sprintf("The clever-tools profile %q expired on %s, run 'clever login' to renew it.", profile.Alias, profile.ExpirationDate),
			)
		}

		clientOptions = append(clientOptions, client.WithOauthConfig(
			creds.consumerKey,
			creds.consumerSecret,
			creds.token,
			creds.secret,
		))

		p.gitAuth = &http.BasicAuth{Username: creds.token, Password: creds.secret}

	} else if !useBearer {
		clientOptions = append(clientOptions, client.WithUserOauthConfig(
			config.Token.ValueString(),
			config.Secret.ValueString(),
		))

		p.gitAuth = &http.BasicAuth{Username: config.Token.ValueString(), Password: config.Secret.ValueString()}
	}

	p.cc = client.New(clientOptions...)

	selfRes := client.Get[map[string]any](ctx, p.cc, "/v2/self")
	if selfRes.HasError() {
		endpoint := config.Endpoint.ValueString()
		tflog.Debug(ctx, fmt.Sprintf("CleverCloud client endpoint=%q", endpoint))

		if selfRes.StatusCode() == 401 || selfRes.StatusCode() == 403 {
			resp.Diagnostics.AddError(
				"CleverCloud authentication failed",
				fmt.Sprintf("Status %d.\n\nCredential priority order:\n1. api_token provider parameter, then CLEVER_API_TOKEN environment variable (Bearer)\n2. token/secret provider parameters (OAuth1)\n3. CC_OAUTH_TOKEN/CC_OAUTH_SECRET environment variables\n4. clever-tools configuration (~/.config/clever-cloud/clever-tools.json)\n\nOriginal error: %s",
					selfRes.StatusCode(), selfRes.Error().Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				"Unknown error from Clever Cloud",
				fmt.Sprintf(
					"Status %d, contact the Clever Cloud support with the next Request ID: '%s'\nError: %s",
					selfRes.StatusCode(), selfRes.SozuID(), selfRes.Error().Error(),
				))
		}
		return
	}

	// We pass the full provider to the children resources
	resp.DataSourceData = p
	resp.ResourceData = p
	resp.ActionData = p

	tflog.Debug(ctx, "provider configured")
}
