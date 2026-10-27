package tmp

import (
	"context"
	"fmt"
	"net/url"

	"go.clever-cloud.dev/client"
)

// The document publishes a `Message` body for these deletions, but the API answers
// with an empty body, which the SDK fails to decode. They move back to the SDK once
// the document is fixed.

// DeleteAddonProvider deletes an addon provider.
func DeleteAddonProvider(ctx context.Context, cc *client.Client, organisationID, providerID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/addonproviders/%s", organisationID, providerID)
	return client.Delete[client.Nothing](ctx, cc, path)
}

// DeleteAddonProviderPlan deletes a plan, by its ID, from an addon provider.
func DeleteAddonProviderPlan(ctx context.Context, cc *client.Client, organisationID, providerID, planID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/addonproviders/%s/plans/%s", organisationID, providerID, planID)
	return client.Delete[client.Nothing](ctx, cc, path)
}

// DeleteTCPRedirection deletes the TCP redirection of an application on a namespace.
func DeleteTCPRedirection(ctx context.Context, cc *client.Client, organisationID, applicationID string, port int64, namespace string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/tcpRedirs/%d?namespace=%s", organisationID, applicationID, port, url.QueryEscape(namespace))
	return client.Delete[client.Nothing](ctx, cc, path)
}
