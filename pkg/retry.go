package pkg

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.clever-cloud.dev/client"
)

// MaxServerErrorRetries bounds how many times a request is sent again after a
// server error.
//
// client.ServerErrorRetryablePolicy has no budget of its own and relies on the
// context to end, but a Terraform operation context is long-lived: an API down
// for good would be retried for as long as the apply runs. The budget therefore
// lives here.
const MaxServerErrorRetries = 2

// RetryServerErrors retries HTTP 5xx up to MaxServerErrorRetries times, waiting
// between attempts and giving up as soon as the context is done.
//
// The API answers 500 often enough that a single one should not fail a plan, an
// apply, or an acceptance run. Client errors are left alone: sending the same
// rejected request again would fail the same way.
func RetryServerErrors(ctx context.Context, req *http.Request, apiErr *client.APIError, retries int) bool {
	if retries >= MaxServerErrorRetries {
		return false
	}

	return client.ServerErrorRetryablePolicy(ctx, req, apiErr, retries)
}

// MaxRateLimitRetries bounds how many times a request the API refused with a
// 429 is sent again.
const MaxRateLimitRetries = 3

// MaxRateLimitWait is the longest the provider waits before one retry. Past it,
// the 429 is reported: some limits ask to wait an hour, which an apply should
// not sit through silently.
const MaxRateLimitWait = 5 * time.Minute

// rateLimitBackoff is the first wait when a 429 does not say how long to wait.
// It doubles on each retry.
const rateLimitBackoff = 2 * time.Second

// RetryPolicy is the retry policy of the provider: rate-limited requests are
// sent again once the API allows it, server errors as RetryServerErrors does.
func RetryPolicy(ctx context.Context, req *http.Request, apiErr *client.APIError, retries int) bool {
	if apiErr != nil && apiErr.StatusCode == http.StatusTooManyRequests {
		return retryRateLimited(ctx, req, apiErr, retries)
	}

	return RetryServerErrors(ctx, req, apiErr, retries)
}

// retryRateLimited waits for as long as a 429 asks, then tells to send the
// request again. A 429 is answered before the request does anything, so every
// method is safe to send again.
//
// Only the status and the headers are read. The body of a 429 is left to the
// client to report: its format is not relied upon.
func retryRateLimited(ctx context.Context, req *http.Request, apiErr *client.APIError, retries int) bool {
	if retries >= MaxRateLimitRetries {
		return false
	}

	wait, told := rateLimitWait(apiErr.Header, time.Now())
	if !told {
		wait = rateLimitBackoff << retries
	}

	fields := map[string]any{"method": req.Method, "path": req.URL.Path, "wait": wait.String(), "retry": retries + 1}
	if wait > MaxRateLimitWait {
		tflog.Warn(ctx, "rate limited by the API for longer than the provider waits, giving up", fields)
		return false
	}
	tflog.Warn(ctx, "rate limited by the API, waiting before sending the request again", fields)

	return sleep(ctx, wait)
}

// rateLimitWait reads how long a rate-limited response asks to wait: from
// Retry-After (RFC 9110, seconds or a date), else from RateLimit-Reset, the
// seconds until the quota resets (IETF draft on RateLimit header fields).
func rateLimitWait(header http.Header, now time.Time) (time.Duration, bool) {
	if value := header.Get("Retry-After"); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
			return time.Duration(seconds) * time.Second, true
		}
		if date, err := http.ParseTime(value); err == nil {
			return max(date.Sub(now), 0), true
		}
	}

	if seconds, err := strconv.Atoi(header.Get("RateLimit-Reset")); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}

	return 0, false
}

// sleep waits for d, and reports false when ctx ended first. A variable so that
// tests do not wait for real.
var sleep = func(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
