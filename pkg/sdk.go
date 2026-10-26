package pkg

import (
	"sync"

	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
	"go.clever-cloud.dev/sdk/models"
	"go.opentelemetry.io/otel"
)

// sdks holds one SDK per client. sdk.NewSDK resolves default credentials from the
// environment and the clever-tools configuration before applying WithClient, which
// is too costly to repeat on every API call.
var sdks sync.Map

// SDK returns the SDK calling the API through cc, built once per client. Resources
// configured through helper.Configurer use its SDK field instead.
func SDK(cc *client.Client) sdk.SDK {
	if s, ok := sdks.Load(cc); ok {
		return s.(sdk.SDK)
	}
	s, _ := sdks.LoadOrStore(cc, sdk.NewSDK(sdk.WithClient(cc)))
	return s.(sdk.SDK)
}

// Deref returns the value p points to, or the zero value of T when p is nil.
// The SDK models an optional field as a pointer where the hand-written client
// used a plain value defaulting to its zero value; Deref keeps that reading.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// Ptr returns a pointer to v, for the SDK's optional request fields.
func Ptr[T any](v T) *T {
	return &v
}

// FlavorName returns the name of a flavor, "" when the flavor or its name is absent.
func FlavorName(flavor *models.Flavor) string {
	if flavor == nil {
		return ""
	}
	return Deref(flavor.Name)
}

// Tracer is the tracer of the SDK, for the few service functions its builder does
// not reach: routes outside the /v2 and /v4 trees, such as /networkgroup/configuration.
var Tracer = otel.Tracer("go.clever-cloud.dev/sdk")
