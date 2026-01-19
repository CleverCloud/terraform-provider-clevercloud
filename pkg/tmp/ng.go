package tmp

import (
	"context"

	"github.com/hashicorp/go-uuid"
	"go.clever-cloud.dev/client"

	"fmt"
)

type Networkgroup struct {
	ID              string   `json:"id"`
	OwnerID         string   `json:"ownerId"`
	Description     *string  `json:"description"`
	NetworkIP       string   `json:"networkIp"`
	LastAllocatedIP string   `json:"lastAllocatedIp"`
	Label           string   `json:"label"`
	Tags            []string `json:"tags"`
	Peers           []Peer   `json:"peers"`
	Members         []Member `json:"members"`
	Version         int64    `json:"version"`
}

type Peer struct {
	ID           string       `json:"id"`
	Label        string       `json:"label"`
	PublicKey    string       `json:"publicKey"`
	Hostname     string       `json:"hostname"`
	ParentMember string       `json:"parentMember"`
	ParentEvent  string       `json:"parentEvent"`
	HV           string       `json:"hv"`
	Endpoint     PeerEndpoint `json:"endpoint"`
}

type PeerEndpoint struct {
	IP string `json:"ngIp"`
}

type Member struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	DomainName string `json:"domainName"`
	Kind       string `json:"kind"`
}

func GetNetworkgroup(ctx context.Context, cc *client.Client, organisationID, networkgroupID string) client.Response[Networkgroup] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s", organisationID, networkgroupID)
	return client.Get[Networkgroup](ctx, cc, path)
}

type NetworkgroupCreation struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Members     []Member `json:"members"`
}

func CreateNetworkgroup(ctx context.Context, cc *client.Client, organisationID string, networkgroup NetworkgroupCreation) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups", organisationID)
	return client.Post[client.Nothing](ctx, cc, path, networkgroup)
}

func DeleteNetworkgroup(ctx context.Context, cc *client.Client, organisationID string, networkgroupID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s", organisationID, networkgroupID)
	return client.Delete[client.Nothing](ctx, cc, path)
}

func GenID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("ng_%s", id)
}

func ListNetworkgroups(ctx context.Context, cc *client.Client, organisationID string) client.Response[[]Networkgroup] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups", organisationID)
	return client.Get[[]Networkgroup](ctx, cc, path)
}

// AddMemberToNetworkgroup adds a member to a networkgroup
func AddMemberToNetworkgroup(ctx context.Context, cc *client.Client, organisationID, networkgroupID string, member Member) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/members", organisationID, networkgroupID)
	return client.Post[client.Nothing](ctx, cc, path, member)
}

// GetMember retrieves a specific member from a networkgroup
func GetMember(ctx context.Context, cc *client.Client, organisationID, networkgroupID, memberID string) client.Response[Member] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/members/%s", organisationID, networkgroupID, memberID)
	return client.Get[Member](ctx, cc, path)
}

// DeleteMember removes a member from a networkgroup
func DeleteMember(ctx context.Context, cc *client.Client, organisationID, networkgroupID, memberID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/members/%s", organisationID, networkgroupID, memberID)
	return client.Delete[client.Nothing](ctx, cc, path)
}

// ListMembers lists all members in a networkgroup
func ListMembers(ctx context.Context, cc *client.Client, organisationID, networkgroupID string) client.Response[[]Member] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/members", organisationID, networkgroupID)
	return client.Get[[]Member](ctx, cc, path)
}
