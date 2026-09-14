package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

type ElasticsearchVersion struct {
	Major int64 `json:"major"`
	Minor int64 `json:"minor"`
	Patch int64 `json:"patch"`
}

// ElasticsearchVersionRequest is a partial version: fields left nil are
// chosen by the API.
type ElasticsearchVersionRequest struct {
	Major *int64 `json:"major"`
	Minor *int64 `json:"minor"`
	Patch *int64 `json:"patch"`
}

type ElasticsearchPlan struct {
	Name     string `json:"name"`
	CPU      int64  `json:"cpu"`
	MemoryMB int64  `json:"memoryMB"`
	DiskMB   int64  `json:"diskMB"`
}

type ElasticsearchNode struct {
	ID   string             `json:"id"`
	Plan *ElasticsearchPlan `json:"plan"`
}

// ElasticsearchClusterDeployed is the status reported by the API once every
// node of the cluster is up.
const ElasticsearchClusterDeployed = "deployed"

type ElasticsearchCluster struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Username         string               `json:"username"`
	Nodes            []ElasticsearchNode  `json:"nodes"`
	Version          ElasticsearchVersion `json:"version"`
	NetworkGroupID   string               `json:"networkGroupId"`
	DeploymentStatus string               `json:"deploymentStatus"`
}

type WannabeElasticsearchCluster struct {
	Name           string                       `json:"name"`
	Version        *ElasticsearchVersionRequest `json:"version"`
	NumberOfNodes  int64                        `json:"numberOfNodes"`
	Plan           string                       `json:"plan"`
	NetworkGroupID *string                      `json:"networkGroupId,omitempty"`
}

// ElasticsearchCredentials is served by the dedicated /credentials endpoint,
// the password is not part of the cluster itself.
type ElasticsearchCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func elasticsearchClustersPath(organisationID string) string {
	return fmt.Sprintf("/v4/elasticsearch/organisations/%s/clusters", organisationID)
}

func elasticsearchClusterPath(organisationID, clusterID string) string {
	return fmt.Sprintf("%s/%s", elasticsearchClustersPath(organisationID), clusterID)
}

// FetchElasticsearchAvailableVersions lists the Elasticsearch versions a
// cluster can be created with.
func FetchElasticsearchAvailableVersions(ctx context.Context, cc *client.Client) client.Response[[]ElasticsearchVersion] {
	return client.Get[[]ElasticsearchVersion](ctx, cc, "/v4/elasticsearch/versions")
}

// FetchElasticsearchAvailablePlans lists the node plans an Elasticsearch
// cluster can be created with.
func FetchElasticsearchAvailablePlans(ctx context.Context, cc *client.Client) client.Response[[]ElasticsearchPlan] {
	return client.Get[[]ElasticsearchPlan](ctx, cc, "/v4/elasticsearch/plans")
}

func CreateElasticsearchCluster(ctx context.Context, cc *client.Client, organisationID string, req WannabeElasticsearchCluster) client.Response[ElasticsearchCluster] {
	return client.Post[ElasticsearchCluster](ctx, cc, elasticsearchClustersPath(organisationID), req)
}

func GetElasticsearchCluster(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[ElasticsearchCluster] {
	return client.Get[ElasticsearchCluster](ctx, cc, elasticsearchClusterPath(organisationID, clusterID))
}

func DeleteElasticsearchCluster(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[client.Nothing] {
	return client.Delete[client.Nothing](ctx, cc, elasticsearchClusterPath(organisationID, clusterID))
}

func GetElasticsearchClusterCredentials(ctx context.Context, cc *client.Client, organisationID, clusterID string) client.Response[ElasticsearchCredentials] {
	return client.Get[ElasticsearchCredentials](ctx, cc, elasticsearchClusterPath(organisationID, clusterID)+"/credentials")
}
