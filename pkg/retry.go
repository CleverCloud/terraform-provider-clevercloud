package pkg

import (
	"context"
	"net/http"

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
