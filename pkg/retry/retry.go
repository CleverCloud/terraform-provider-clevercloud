package retry

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

// Config holds retry configuration
type Config struct {
	// MaxAttempts is the maximum number of attempts (including the initial request)
	MaxAttempts int
	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the exponential backoff multiplier
	Multiplier float64
}

// DefaultConfig returns sensible defaults for retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// shouldRetry determines if a response should be retried based on status code
func shouldRetry(statusCode int) bool {
	// Retry on 503 Service Unavailable
	return statusCode == 503
}

// calculateBackoff calculates the backoff duration for a given attempt
func calculateBackoff(attempt int, config Config) time.Duration {
	// Calculate exponential backoff: initialDelay * multiplier^attempt
	backoff := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at maxDelay
	if backoff > float64(config.MaxDelay) {
		backoff = float64(config.MaxDelay)
	}

	return time.Duration(backoff)
}

// WithRetry wraps a client.Response-returning function with retry logic
// It will retry on 503 errors with exponential backoff
func WithRetry[T any](
	ctx context.Context,
	operation func() client.Response[T],
	operationName string,
) client.Response[T] {
	return WithRetryConfig(ctx, operation, operationName, DefaultConfig())
}

// WithRetryConfig wraps a client.Response-returning function with retry logic using custom config
func WithRetryConfig[T any](
	ctx context.Context,
	operation func() client.Response[T],
	operationName string,
	config Config,
) client.Response[T] {
	var lastResponse client.Response[T]

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute the operation
		lastResponse = operation()

		// If successful or non-retryable error, return immediately
		if !lastResponse.HasError() || !shouldRetry(lastResponse.StatusCode()) {
			if attempt > 0 {
				tflog.Debug(ctx, fmt.Sprintf("%s succeeded after %d retries", operationName, attempt))
			}
			return lastResponse
		}

		// Check if we should retry
		if attempt < config.MaxAttempts-1 {
			backoff := calculateBackoff(attempt, config)
			tflog.Warn(ctx, fmt.Sprintf(
				"%s failed with status %d (attempt %d/%d), retrying in %v",
				operationName,
				lastResponse.StatusCode(),
				attempt+1,
				config.MaxAttempts,
				backoff,
			))

			// Wait before retrying
			select {
			case <-ctx.Done():
				tflog.Error(ctx, fmt.Sprintf("%s cancelled during retry backoff", operationName))
				return lastResponse
			case <-time.After(backoff):
				// Continue to next attempt
			}
		} else {
			tflog.Error(ctx, fmt.Sprintf(
				"%s failed after %d attempts, giving up",
				operationName,
				config.MaxAttempts,
			))
		}
	}

	return lastResponse
}
