package impl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

// detectCredentialIssue analyzes the credential context and returns an appropriate error message
func detectCredentialIssue() string {
	// Check environment variables first (highest priority)
	ccToken := os.Getenv("CC_OAUTH_TOKEN")
	ccSecret := os.Getenv("CC_OAUTH_SECRET")

	if ccToken == "" && ccSecret == "" {
		// Check if clever-tools config exists
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".config", "clever-cloud", "clever-tools.json")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return "No CleverCloud credentials found"
			}
			return ""
		}

		// Fallback message
		return "Cannot access home directory to check credentials"
	}

	if ccToken == "" || ccSecret == "" {
		return "CC_OAUTH_TOKEN and CC_OAUTH_SECRET environment variables must both be set"
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
		clientOptions = append(clientOptions, client.WithAutoOauthConfig())

		tmpClient := client.New()
		if err := detectCredentialIssue(); err != "" {
			resp.Diagnostics.AddError(
				"Invalid CleverCloud authentication configuration",
				err,
			)
			return
		}
		c := tmpClient.GuessOauth1Config()

		// Check if GuessOauth1Config returned nil (invalid/missing credentials)
		if c == nil {
			resp.Diagnostics.AddError(
				"CleverCloud authentication empty",
				"Something went wrong while trying to guess OAuth1 credentials",
			)
			return
		}

		p.gitAuth = &http.BasicAuth{Username: c.AccessToken, Password: c.AccessSecret}

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

	tflog.Debug(ctx, "provider configured")
}
