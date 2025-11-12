package tmp

import (
	"context"
	"net/url"

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

func SearchNetworkgroup(ctx context.Context, cc *client.Client, organisationID, query string) client.Response[[]any] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/search?query=%s", organisationID, url.QueryEscape(query))
	return client.Get[[]any](ctx, cc, path)
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

// UpdateNetworkgroup updates an existing networkgroup
func UpdateNetworkgroup(ctx context.Context, cc *client.Client, organisationID, networkgroupID string, networkgroup Networkgroup) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s", organisationID, networkgroupID)
	return client.Put[client.Nothing](ctx, cc, path, networkgroup)
}

// PeerCreation represents a peer to be added to a networkgroup
type PeerCreation struct {
	Label     string       `json:"label"`
	PublicKey string       `json:"publicKey"`
	Hostname  string       `json:"hostname"`
	Endpoint  PeerEndpoint `json:"endpoint"`
}

// PeerCreatedResponse represents the response when a peer is created
type PeerCreatedResponse struct {
	PeerID string `json:"peerId"`
}

// AddPeerToNetworkgroup adds a peer to a networkgroup
func AddPeerToNetworkgroup(ctx context.Context, cc *client.Client, organisationID, networkgroupID string, peer PeerCreation) client.Response[PeerCreatedResponse] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/peers", organisationID, networkgroupID)
	return client.Post[PeerCreatedResponse](ctx, cc, path, peer)
}

// GetPeer retrieves a specific peer from a networkgroup
func GetPeer(ctx context.Context, cc *client.Client, organisationID, networkgroupID, peerID string) client.Response[Peer] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/peers/%s", organisationID, networkgroupID, peerID)
	return client.Get[Peer](ctx, cc, path)
}

// DeletePeer removes a peer from a networkgroup
func DeletePeer(ctx context.Context, cc *client.Client, organisationID, networkgroupID, peerID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/peers/%s", organisationID, networkgroupID, peerID)
	return client.Delete[client.Nothing](ctx, cc, path)
}

// ListPeers lists all peers in a networkgroup
func ListPeers(ctx context.Context, cc *client.Client, organisationID, networkgroupID string) client.Response[[]Peer] {
	path := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/peers", organisationID, networkgroupID)
	return client.Get[[]Peer](ctx, cc, path)
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
