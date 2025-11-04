package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

type (
	// LoadBalancer is a minimal representation of a Clever Cloud load balancer as returned by the v4 API.
	// Fields not represented here will be ignored by the JSON unmarshaler.
	LoadBalancer struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		ZoneID string          `json:"zoneId"`
		DNS    LoadBalancerDNS `json:"dns"`
	}

	// LoadBalancerDNS represents the DNS configuration of a load balancer
	LoadBalancerDNS struct {
		CNAME string   `json:"cname"`
		A     []string `json:"a"`
	}

	LoadBalancers []LoadBalancer
)

// GetLoadBalancer retrieves the default load balancer for a given organisation and application.
// GET /v4/load-balancers/organisations/{organisationId}/applications/{applicationId}/load-balancers/default
func GetLoadBalancer(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[LoadBalancers] {
	path := fmt.Sprintf("/v4/load-balancers/organisations/%s/applications/%s/load-balancers/default", organisationID, applicationID)
	return client.Get[LoadBalancers](ctx, cc, path)
}
