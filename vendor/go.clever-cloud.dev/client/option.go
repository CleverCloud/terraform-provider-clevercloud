package client

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// Set API endoint, default: API_ENDPOINT.
func WithEndpoint(endpoint string) func(*Client) {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// Set a logger, default: discard.
func WithLogger(logger logrus.FieldLogger) func(*Client) {
	return func(c *Client) {
		c.log = logger
	}
}

// Set custom http client, default: http.DefaultClient.
func WithHTTPClient(httpClient *http.Client) func(*Client) {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// Set OAuth1 credentials, default: none.
func WithOauthConfig(consumerKey, consumerSecret, accessToken, accessSecret string) func(*Client) {
	return func(c *Client) {
		c.authenticator = &OAuth1Config{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			AccessToken:    accessToken,
			AccessSecret:   accessSecret,
		}
	}
}

// Set OAuth1 user credentials, default: none.
func WithUserOauthConfig(accessToken, accessSecret string) func(*Client) {
	return func(c *Client) {
		c.authenticator = &OAuth1Config{
			ConsumerKey:    OAUTH_CONSUMER_KEY,
			ConsumerSecret: OAUTH_CONSUMER_SECRET,
			AccessToken:    accessToken,
			AccessSecret:   accessSecret,
		}
	}
}

// Set OAuth1 credentials from environment, default: none.
func WithAutoOauthConfig() func(*Client) {
	return func(c *Client) {
		conf := c.GuessOauth1Config()
		if conf == nil {
			return
		}

		if conf.ConsumerKey == "" {
			conf.ConsumerKey = OAUTH_CONSUMER_KEY
		}

		if conf.ConsumerSecret == "" {
			conf.ConsumerSecret = OAUTH_CONSUMER_SECRET
		}

		c.authenticator = conf
	}
}

// Set OAuth1 credentials from environment, default: none.
func WithAutoAuthConfig() func(*Client) {
	return func(c *Client) {
		WithAutoOauthConfig()(c)

		if c.authenticator == nil {
			WithEndpoint(BRIDGE_API_ENDPOINT)(c)
			WithBearerAuth("")(c)
		}
	}
}

func WithBearerAuth(token string) func(c *Client) {
	return func(c *Client) {
		if c.endpoint == API_ENDPOINT {
			c.log.Warnf("bearer tokens need an alternative endpoint, consider using '%s'", BRIDGE_API_ENDPOINT)
		}

		if token == "" {
			c.authenticator = c.guessBearerConfigFromEnv()
		} else {
			c.authenticator = &BearerConfig{Token: token}
		}
	}
}
