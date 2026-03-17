package tmp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.clever-cloud.dev/client"
)

type (
	DRAIN_KIND string

	// Drain is a minimal representation of a Clever Cloud drain as returned by the v4 API.
	// Fields not represented here will be ignored by the JSON unmarshaler.
	Drain struct {
		ID            string          `json:"id"`
		OwnerID       string          `json:"tenantId"`
		ApplicationID string          `json:"resourceId"`
		Kind          DRAIN_KIND      `json:"kind"`
		Status        DrainStatus     `json:"status"`
		Recipient     json.RawMessage `json:"recipient"`
		// backlog
		// execution
	}

	// DrainStatus represents the status object returned by the API
	DrainStatus struct {
		ID       string `json:"id"`
		DrainID  string `json:"drainId"`
		Status   string `json:"status"`
		Date     string `json:"date"`
		AuthorID string `json:"authorId,omitempty"`
	}
	Drains []Drain

	RecipientDatadog struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	RecipientElasticsearch struct {
		Type            string `json:"type"`
		URL             string `json:"url"`
		Username        string `json:"username"`
		Password        string `json:"password"`
		Index           string `json:"index"`
		TLSVerification string `json:"tlsVerification"` // DEFAULT | TRUSTFUL
	}

	RecipientNewrelic struct {
		Type   string `json:"type"`
		URL    string `json:"url"`    // e.g. https://log-api.newrelic.com/log/v1 or https://log-api.eu.newrelic.com/log/v1
		APIKey string `json:"apiKey"` // New Relic API key used for ingestion
	}

	RecipientOVH struct {
		Type                        string `json:"type"`
		URL                         string `json:"url"`
		Token                       string `json:"token"`
		RFC5424StructuredDataParams string `json:"rfc5424StructuredDataParameters"`
	}

	RecipientRaw struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	RecipientSyslogUDP struct {
		Type  string `json:"type"`
		URL   string `json:"url"`
		Token string `json:"token"`
	}

	RecipientSyslogTCP struct {
		Type  string `json:"type"`
		URL   string `json:"url"`
		Token string `json:"token"`
	}
)

const (
	DRAIN_KIND_LOG       DRAIN_KIND = "LOG"
	DRAIN_KIND_ACCESSLOG DRAIN_KIND = "ACCESSLOG"
	DRAIN_KIND_AUDITLOG  DRAIN_KIND = "AUDITLOG"
)

// WannabeDrain is the minimal payload to create a drain.
// It primarily carries the recipient configuration and optional kind.
type WannabeDrain struct {
	Recipient json.RawMessage `json:"recipient"`
	Kind      DRAIN_KIND      `json:"kind,omitempty"`
}

// retryOn503 wraps an API call and retries it on 503 errors with exponential backoff.
// It will retry up to 5 times with delays of 2s, 4s, 8s, 16s, 32s.
func retryOn503[T any](ctx context.Context, fn func() client.Response[T]) client.Response[T] {
	maxRetries := 5
	baseDelay := 2 * time.Second

	var response client.Response[T]
	for attempt := 0; attempt < maxRetries; attempt++ {
		response = fn()

		// If no error or not a 503, return immediately
		if !response.HasError() || response.StatusCode() != 503 {
			return response
		}

		// If this is the last attempt, return the error
		if attempt == maxRetries-1 {
			return response
		}

		// Wait before retrying with exponential backoff
		delay := baseDelay * time.Duration(1<<uint(attempt))
		select {
		case <-ctx.Done():
			return response
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	return response
}

// CreateDrain creates a drain for a given organisation and application.
// POST /v4/drains/organisations/{ownerId}/applications/{applicationId}/drains
// Automatically retries on 503 errors with exponential backoff.
func CreateDrain(ctx context.Context, cc *client.Client, organisationID, applicationID string, req WannabeDrain) client.Response[Drain] {
	path := fmt.Sprintf("/v4/drains/organisations/%s/applications/%s/drains", organisationID, applicationID)
	return retryOn503(ctx, func() client.Response[Drain] {
		return client.Post[Drain](ctx, cc, path, req)
	})
}

// GetDrain retrieves a specific drain.
// GET /v4/drains/organisations/{ownerId}/applications/{applicationId}/drains/{drainId}
// Automatically retries on 503 errors with exponential backoff.
func GetDrain(ctx context.Context, cc *client.Client, organisationID, applicationID, drainID string) client.Response[Drain] {
	path := fmt.Sprintf("/v4/drains/organisations/%s/applications/%s/drains/%s", organisationID, applicationID, drainID)
	return retryOn503(ctx, func() client.Response[Drain] {
		return client.Get[Drain](ctx, cc, path)
	})
}

// DeleteDrain deletes a specific drain.
// DELETE /v4/drains/organisations/{ownerId}/applications/{applicationId}/drains/{drainId}
// The API returns the deleted Drain in the response body (HTTP 200).
// Automatically retries on 503 errors with exponential backoff.
func DeleteDrain(ctx context.Context, cc *client.Client, organisationID, applicationID, drainID string) client.Response[Drain] {
	path := fmt.Sprintf("/v4/drains/organisations/%s/applications/%s/drains/%s", organisationID, applicationID, drainID)
	return retryOn503(ctx, func() client.Response[Drain] {
		return client.Delete[Drain](ctx, cc, path)
	})
}
