package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

// AddAppLinkedAddons links an add-on to an application. The document publishes the
// body as `text/plain`, and the SDK only sends JSON bodies; the API takes the add-on
// ID as a JSON string.
func AddAppLinkedAddons(ctx context.Context, cc *client.Client, organisationID, applicationID, addonID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/addons", organisationID, applicationID)
	return client.Post[client.Nothing](ctx, cc, path, addonID)
}
