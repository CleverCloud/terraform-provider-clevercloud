package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

// Tags are free-form strings, by convention either a plain word ("prod") or a
// key/value pair ("env:prod"). The API does not validate them.

// GetAppTags returns the tags of an application
func GetAppTags(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[[]string] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/tags", organisationID, applicationID)
	return client.Get[[]string](ctx, cc, path)
}

// UpdateAppTags replaces all the tags of an application
func UpdateAppTags(ctx context.Context, cc *client.Client, organisationID, applicationID string, tags []string) client.Response[[]string] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/tags", organisationID, applicationID)
	return client.Put[[]string](ctx, cc, path, tags)
}

// GetAddonTags returns the tags of an add-on. addonID is the "addon_xxx" ID,
// the API does not resolve real IDs on this route.
func GetAddonTags(ctx context.Context, cc *client.Client, organisationID, addonID string) client.Response[[]string] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s/tags", organisationID, addonID)
	return client.Get[[]string](ctx, cc, path)
}

// UpdateAddonTags replaces all the tags of an add-on. addonID is the
// "addon_xxx" ID, the API does not resolve real IDs on this route.
func UpdateAddonTags(ctx context.Context, cc *client.Client, organisationID, addonID string, tags []string) client.Response[[]string] {
	path := fmt.Sprintf("/v2/organisations/%s/addons/%s/tags", organisationID, addonID)
	return client.Put[[]string](ctx, cc, path, tags)
}
