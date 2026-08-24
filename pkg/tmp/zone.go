package tmp

import (
	"context"

	"go.clever-cloud.dev/client"
)

type (
	// Zone is a minimal representation of a Clever Cloud zone as returned by the v2 API.
	// Fields not represented here will be ignored by the JSON unmarshaler.
	Zone struct {
		Name                string `json:"name"`
		Internal            bool   `json:"internal"`
		CorrespondingRegion string `json:"correspondingRegion"`
	}

	Zones []Zone
)

// GetZones lists the zones the API knows about.
// GET /v2/products/zones
func GetZones(ctx context.Context, cc *client.Client) client.Response[Zones] {
	return client.Get[Zones](ctx, cc, "/v2/products/zones")
}
