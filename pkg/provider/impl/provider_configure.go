package impl

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

type oauthCreds struct{ consumerKey, consumerSecret, token, secret string }

// resolveCreds mirrors go.clever-cloud.dev/client credential precedence (env, then
// clever-tools config).
func resolveCreds() (*oauthCreds, *client.Profile, diag.Diagnostics) {
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

	path := client.ConfigFilePath()

	profile, err := client.ActiveProfile(path)
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

	// Allow to get creds from CLI config directory or by injected variables
	var clientOptions []func(*client.Client)
	if !config.Endpoint.IsUnknown() && !config.Endpoint.IsNull() && config.Endpoint.ValueString() != "" {
		clientOptions = append(clientOptions, client.WithEndpoint(config.Endpoint.ValueString()))
	}

	// New branch: allow setting all OAuth1 params
	if !config.ConsumerKey.IsUnknown() && !config.ConsumerKey.IsNull() && config.ConsumerKey.ValueString() != "" &&
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

	} else if config.Secret.IsUnknown() ||
		config.Token.IsUnknown() ||
		config.Secret.IsNull() ||
		config.Token.IsNull() {
		creds, profile, diags := resolveCreds()
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if creds == nil {
			searched := client.ConfigFilePath()
			if searched == "" {
				searched = "no clever-tools configuration file found"
			}

			resp.Diagnostics.AddError(
				"CleverCloud authentication empty",
				fmt.Sprintf(
					"No credentials found (%s).\n\nEither set the CC_OAUTH_TOKEN and CC_OAUTH_SECRET environment variables, run 'clever login', or set the token and secret provider parameters.",
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

	} else {
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
				fmt.Sprintf("Status %d.\n\nCredential priority order:\n1. CC_OAUTH_TOKEN/CC_OAUTH_SECRET environment variables\n2. clever-tools configuration (~/.config/clever-cloud/clever-tools.json)\n3. Terraform provider token/secret parameters\n\nOriginal error: %s",
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
	resp.StateStoreData = p

	tflog.Debug(ctx, "provider configured")
}
